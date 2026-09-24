package service

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/Nurlan270/cloud-storage-go/internal/cloud_storage/transport/http/dto/request"
	"github.com/Nurlan270/cloud-storage-go/internal/cloud_storage/transport/http/dto/response"
	corectx "github.com/Nurlan270/cloud-storage-go/internal/core/context"
	errs "github.com/Nurlan270/cloud-storage-go/internal/core/errors"
	"github.com/Nurlan270/cloud-storage-go/internal/core/logger"
	"github.com/Nurlan270/cloud-storage-go/internal/core/minio"
	"github.com/Nurlan270/cloud-storage-go/internal/core/models"
)

type DirectoryService interface {
	Create(ctx context.Context, req request.CreateDirectory) (response.ResourceInfo, error)
	GetContent(ctx context.Context, req request.GetDirectoryContent) ([]models.Resource, error)
}

type DirectoryRepository interface {
	GetAll(ctx context.Context, dir models.Resource, recursive bool) ([]models.Resource, error)
	Create(ctx context.Context, dir models.Resource) (models.Resource, error)
	Exists(ctx context.Context, dir models.Resource) (bool, error)
}

type directoryService struct {
	client  minio.Client
	pool    *pgxpool.Pool
	dirRepo DirectoryRepository

	log *zap.Logger
}

func NewDirectoryService(
	client minio.Client,
	pool *pgxpool.Pool,
	dirRepo DirectoryRepository,
) DirectoryService {
	log := logger.Get().With(
		zap.String("src", "directory service"))

	return &directoryService{
		client:  client,
		pool:    pool,
		dirRepo: dirRepo,
		log:     log,
	}
}

func (s *directoryService) Create(
	ctx context.Context,
	req request.CreateDirectory,
) (response.ResourceInfo, error) {
	//	Start TX
	tx, txErr := s.pool.Begin(ctx)
	if txErr != nil {
		s.log.Error("tx: failed to start", zap.Error(txErr))
		return response.ResourceInfo{}, txErr
	}
	defer tx.Rollback(ctx)

	//	Put TX into ctx
	ctx = corectx.NewTxContext(ctx, tx)

	user := corectx.UserFromContext(ctx)
	path, name := splitPath(req.Path)
	dir := models.Resource{
		UserID: user.ID,
		Path:   path,
		Name:   name,
		Type:   models.TypeDir,
	}

	parentPath, parentName := splitPath(path)
	parentDir := models.Resource{
		UserID: user.ID,
		Path:   parentPath,
		Name:   parentName,
		Type:   models.TypeDir,
	}

	//	Check whether parent directory exists
	if exists, err := s.dirRepo.Exists(ctx, parentDir); err != nil {
		return response.ResourceInfo{}, err
	} else if !exists {
		return response.ResourceInfo{}, errs.ErrParentDirectoryNotFound
	}

	//	Put into DB
	res, err := s.dirRepo.Create(ctx, dir)
	if err != nil {
		return response.ResourceInfo{}, err
	}

	putOpts := minio.PutOptions{
		Reader:      nil,
		Key:         dir.ObjectKey(),
		Size:        0,
		ContentType: "application/octet-stream",
	}

	//	Put into Bucket
	if err = s.client.Put(ctx, putOpts); err != nil {
		return response.ResourceInfo{}, err
	}

	//	Commit TX
	if err = tx.Commit(ctx); err != nil {
		s.log.Error("tx: failed to commit", zap.Error(err))
		return response.ResourceInfo{}, err
	}

	return response.ResourceInfo{
		Path: res.Path,
		Name: res.Name,
		Type: res.Type,
	}, nil
}

func (s *directoryService) GetContent(
	ctx context.Context,
	req request.GetDirectoryContent,
) ([]models.Resource, error) {
	user := corectx.UserFromContext(ctx)

	path, name := splitPath(req.Path)

	dir := models.Resource{
		UserID: user.ID,
		Path:   path,
		Name:   name,
		Type:   models.TypeDir,
	}

	//	Get from DB
	resources, err := s.dirRepo.GetAll(ctx, dir, false)
	if err != nil {
		return nil, err
	}

	return resources, nil
}
