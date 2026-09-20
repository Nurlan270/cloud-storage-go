package service

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/minio/minio-go/v7"
	"go.uber.org/zap"

	"github.com/Nurlan270/cloud-storage-go/internal/cloud_storage/transport/http/dto/request"
	"github.com/Nurlan270/cloud-storage-go/internal/cloud_storage/transport/http/dto/response"
	corectx "github.com/Nurlan270/cloud-storage-go/internal/core/context"
	errs "github.com/Nurlan270/cloud-storage-go/internal/core/errors"
	"github.com/Nurlan270/cloud-storage-go/internal/core/logger"
	"github.com/Nurlan270/cloud-storage-go/internal/core/models"
)

type DirectoryService interface {
	Create(ctx context.Context, req request.CreateDirectory) (response.ResourceInfo, error)
	GetContent(ctx context.Context, req request.GetDirectoryContent) ([]models.Resource, error)
}

type DirectoryRepository interface {
	GetAll(ctx context.Context, userID uint64, path string) ([]models.Resource, error)
	Create(ctx context.Context, dir models.Resource) (models.Resource, error)
	Exists(ctx context.Context, dir models.Resource) (bool, error)
}

type directoryService struct {
	client  *minio.Client
	pool    *pgxpool.Pool
	dirRepo DirectoryRepository

	log *logger.Logger
}

func NewDirectoryService(
	client *minio.Client,
	pool *pgxpool.Pool,
	dirRepo DirectoryRepository,
) DirectoryService {
	return &directoryService{
		client:  client,
		pool:    pool,
		dirRepo: dirRepo,
		log:     logger.Get(),
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
	if parentDir.Path != "/" && parentDir.Name != "" {
		if exists, err := s.dirRepo.Exists(ctx, parentDir); err != nil {
			s.log.Error("directory repo: failed to check if directory exists", zap.Error(err))
			return response.ResourceInfo{}, err
		} else if !exists {
			return response.ResourceInfo{}, errs.ErrParentDirectoryNotExists
		}
	}

	//	Put into DB
	res, err := s.dirRepo.Create(ctx, dir)
	if err != nil {
		if !errors.Is(err, errs.ErrDirectoryAlreadyExists) {
			s.log.Error("directory repo: failed to get dir", zap.Error(err))
		}

		return response.ResourceInfo{}, err
	}

	//	Put into Bucket
	if _, err = s.client.PutObject(
		ctx,
		Bucket,
		dir.ObjectKey(),
		nil,
		0,
		minio.PutObjectOptions{},
	); err != nil {
		s.log.Error("failed to put object into bucket", zap.Error(err))
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

	path := "/"
	if req.Path != "/" {
		path = strings.Trim(req.Path, "/") + "/"
	}

	//	Get from DB
	list, err := s.dirRepo.GetAll(ctx, user.ID, path)
	if err != nil {
		if !errors.Is(err, errs.ErrDirectoryNotFound) {
			s.log.Error("directory repo: failed to get dir", zap.Error(err))
		}

		return nil, err
	}

	return list, nil
}
