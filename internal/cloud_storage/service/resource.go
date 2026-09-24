package service

import (
	"context"
	"io"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/Nurlan270/cloud-storage-go/internal/cloud_storage/transport/http/dto/request"
	"github.com/Nurlan270/cloud-storage-go/internal/cloud_storage/transport/http/dto/response"
	corectx "github.com/Nurlan270/cloud-storage-go/internal/core/context"
	"github.com/Nurlan270/cloud-storage-go/internal/core/logger"
	"github.com/Nurlan270/cloud-storage-go/internal/core/minio"
	"github.com/Nurlan270/cloud-storage-go/internal/core/models"
	"github.com/Nurlan270/cloud-storage-go/internal/core/util"
)

type ResourceService interface {
	Upload(ctx context.Context, req request.UploadResource) (response.ResourceInfoList, error)
	GetInfo(ctx context.Context, req request.GetResourceInfo) (response.ResourceInfo, error)
	Search(ctx context.Context, req request.SearchResource) (response.ResourceInfoList, error)
	Delete(ctx context.Context, req request.DeleteResource) error
	Download(ctx context.Context, req request.DownloadResource) (DownloadResult, error)
}

type ResourceRepository interface {
	BatchCreate(ctx context.Context, resources []models.Resource) error
	Get(ctx context.Context, resource *models.Resource) (*models.Resource, error)
	Update(ctx context.Context, old *models.Resource, new *models.Resource) (*models.Resource, error)
	Search(ctx context.Context, userID uint64, query string) ([]*models.Resource, error)
	Delete(ctx context.Context, resource *models.Resource) error
}

type MinioClient interface {
	Get(ctx context.Context, resource models.Resource) (minio.GetResult, error)
	Put(ctx context.Context, opts minio.PutOptions) error
	PutAll(ctx context.Context, entities []minio.PutAllEntities) error
	Delete(ctx context.Context, resource models.Resource) error
	DeleteAll(ctx context.Context, resources []models.Resource) error
}

type resourceService struct {
	client       MinioClient
	pool         *pgxpool.Pool
	resourceRepo ResourceRepository
	dirRepo      DirectoryRepository

	log *zap.Logger
}

func NewResourceService(
	client MinioClient,
	pool *pgxpool.Pool,
	resourceRepo ResourceRepository,
	dirRepo DirectoryRepository,
) ResourceService {
	log := logger.Get().With(
		zap.String("src", "resource service"))

	return &resourceService{
		client:       client,
		pool:         pool,
		resourceRepo: resourceRepo,
		dirRepo:      dirRepo,
		log:          log,
	}
}

func (s *resourceService) Upload(
	ctx context.Context,
	req request.UploadResource,
) (response.ResourceInfoList, error) {
	user := corectx.UserFromContext(ctx)

	uploadEntities := make([]minio.PutAllEntities, 0, len(req.Object))
	rawResourceList := make([]models.Resource, 0, len(req.Object))

	for _, obj := range req.Object {
		fullPath, err := getFullPath(obj, req.Path)
		if err != nil {
			return nil, err
		}

		resources := buildResourcesFromPath(user.ID, fullPath, &obj.Size)

		rawResourceList = append(rawResourceList, resources...)

		for _, resource := range resources {
			if resource.IsDir() {
				continue
			}

			uploadEntities = append(uploadEntities, minio.PutAllEntities{
				Resource: resource,
				Object:   obj,
			})
		}
	}

	//	fixme: this needs optimization, it does unnecessary work
	//	Drop all duplicate values from slice
	resourceList := util.UniqueSlice[models.Resource](rawResourceList)

	//	Begin TX
	tx, txErr := s.pool.Begin(ctx)
	if txErr != nil {
		s.log.Error("tx: failed to start", zap.Error(txErr))
		return nil, txErr
	}
	defer tx.Rollback(ctx)

	//	Put TX into ctx
	ctx = corectx.NewTxContext(ctx, tx)

	//	Bulk insert resources into DB
	if err := s.resourceRepo.BatchCreate(ctx, resourceList); err != nil {
		return nil, err
	}

	//	Put resources into bucket
	if err := s.client.PutAll(ctx, uploadEntities); err != nil {
		return response.ResourceInfoList{}, err
	}

	//	Commit TX
	if err := tx.Commit(ctx); err != nil {
		s.log.Error("tx: failed to commit", zap.Error(err))
		return nil, err
	}

	//	Convert models into response dto
	resp := make(response.ResourceInfoList, 0, len(uploadEntities))
	for _, res := range resourceList {
		resp = append(resp, &response.ResourceInfo{
			Path: res.Path,
			Name: res.Name,
			Size: res.Size,
			Type: res.Type,
		})
	}

	return resp, nil
}

func (s *resourceService) GetInfo(
	ctx context.Context,
	req request.GetResourceInfo,
) (response.ResourceInfo, error) {
	user := corectx.UserFromContext(ctx)

	path, name := splitPath(req.Path)
	resourceType := getResourceType(req.Path)
	resource := &models.Resource{
		UserID: user.ID,
		Path:   path,
		Name:   name,
		Type:   resourceType,
	}

	res, err := s.resourceRepo.Get(ctx, resource)
	if err != nil {
		return response.ResourceInfo{}, err
	}

	return response.ResourceInfo{
		Path: res.Path,
		Name: res.Name,
		Size: res.Size,
		Type: res.Type,
	}, nil
}

func (s *resourceService) Search(
	ctx context.Context,
	req request.SearchResource,
) (response.ResourceInfoList, error) {
	user := corectx.UserFromContext(ctx)

	list, err := s.resourceRepo.Search(ctx, user.ID, req.Query)
	if err != nil {
		return nil, err
	}

	//	fixme: figure out how to optimize this
	info := make(response.ResourceInfoList, 0, len(list))
	for _, resource := range list {
		info = append(info, &response.ResourceInfo{
			Path: resource.Path,
			Name: resource.Name,
			Size: resource.Size,
			Type: resource.Type,
		})
	}

	return info, nil
}

//nolint:gocyclo
func (s *resourceService) Delete(ctx context.Context, req request.DeleteResource) error {
	user := corectx.UserFromContext(ctx)

	path, name := splitPath(req.Path)
	resourceType := getResourceType(req.Path)

	//	Start TX
	tx, txErr := s.pool.Begin(ctx)
	if txErr != nil {
		s.log.Error("tx: failed to start", zap.Error(txErr))
		return txErr
	}
	defer tx.Rollback(ctx)

	//	Put TX into ctx
	ctx = corectx.NewTxContext(ctx, tx)

	resource := models.Resource{
		UserID: user.ID,
		Path:   path,
		Name:   name,
		Type:   resourceType,
	}

	var err error

	resources := make([]models.Resource, 0)

	if resource.IsDir() {
		//	Get directory content
		resources, err = s.dirRepo.GetAll(ctx, resource, true)
		if err != nil {
			return err
		}
	}

	//	Delete from Bucket
	if len(resources) > 0 {
		if err = s.client.DeleteAll(ctx, resources); err != nil {
			return err
		}
	} else {
		if err = s.client.Delete(ctx, resource); err != nil {
			return err
		}
	}

	//	Delete from DB
	if err = s.resourceRepo.Delete(ctx, &resource); err != nil {
		return err
	}

	//	Commit TX
	if err = tx.Commit(ctx); err != nil {
		s.log.Error("tx: failed to commit", zap.Error(err))
		return err
	}

	return nil
}

type DownloadResult struct {
	Content io.ReadCloser
	Name    string
	Size    int64
	Type    string
}

func (s *resourceService) Download(
	ctx context.Context,
	req request.DownloadResource,
) (DownloadResult, error) {
	//	Build resource from request
	user := corectx.UserFromContext(ctx)
	path, name := splitPath(req.Path)
	resourceType := getResourceType(req.Path)

	resource := models.Resource{
		UserID: user.ID,
		Path:   path,
		Name:   name,
		Type:   resourceType,
	}

	//	Check whether provided resource exists
	if _, err := s.resourceRepo.Get(ctx, &resource); err != nil {
		return DownloadResult{}, err
	}

	//	Get actual resource from MinIO
	result, err := s.client.Get(ctx, resource)
	if err != nil {
		return DownloadResult{}, err
	}

	return DownloadResult{
		Content: result.Content,
		Name:    result.Name,
		Size:    result.Size,
		Type:    result.ContentType,
	}, nil
}
