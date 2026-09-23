package service

import (
	"context"
	"errors"
	"io"
	"mime/multipart"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/minio/minio-go/v7"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/Nurlan270/cloud-storage-go/internal/cloud_storage/transport/http/dto/request"
	"github.com/Nurlan270/cloud-storage-go/internal/cloud_storage/transport/http/dto/response"
	corectx "github.com/Nurlan270/cloud-storage-go/internal/core/context"
	errs "github.com/Nurlan270/cloud-storage-go/internal/core/errors"
	"github.com/Nurlan270/cloud-storage-go/internal/core/logger"
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

type resourceService struct {
	client       *minio.Client
	pool         *pgxpool.Pool
	resourceRepo ResourceRepository
	dirRepo      DirectoryRepository

	log *logger.Logger
}

func NewResourceService(
	client *minio.Client,
	pool *pgxpool.Pool,
	resourceRepo ResourceRepository,
	dirRepo DirectoryRepository,
) ResourceService {
	return &resourceService{
		client:       client,
		pool:         pool,
		resourceRepo: resourceRepo,
		dirRepo:      dirRepo,
		log:          logger.Get(),
	}
}

const Bucket = "user-files"

type UploadResource struct {
	Resource models.Resource
	Object   *multipart.FileHeader
}

//nolint:gocyclo
func (s *resourceService) Upload(
	ctx context.Context,
	req request.UploadResource,
) (response.ResourceInfoList, error) {
	user := corectx.UserFromContext(ctx)

	uploadResources := make([]UploadResource, 0, len(req.Object))
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

			uploadResources = append(uploadResources, UploadResource{
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
		if !errors.Is(err, errs.ErrResourceAlreadyExists) {
			s.log.Error("resource repo: failed to bulk insert", zap.Error(err))
		}

		return nil, err
	}

	//	MinIO upload
	g, gctx := errgroup.WithContext(ctx)

	objsCh := make(chan minio.SnowballObject)

	g.Go(func() error {
		defer close(objsCh)

		for _, item := range uploadResources {
			file, err := item.Object.Open()
			if err != nil {
				s.log.Error(
					"failed to open object",
					zap.Any("object", item.Object),
					zap.Error(err),
				)

				return err
			}

			//	File closer
			closeFileFn := func() {
				if err = file.Close(); err != nil {
					s.log.Warn("failed to close file",
						zap.Any("file", file), zap.Error(err),
					)
				}
			}

			obj := minio.SnowballObject{
				Key:     item.Resource.ObjectKey(),
				Size:    *item.Resource.Size,
				Content: file,
				Close:   closeFileFn,
			}

			select {
			case objsCh <- obj:
				// MinIO consumed object. It's in charge of closing opened file.
			case <-gctx.Done():
				// MinIO stopped consuming. Close opened file manually.
				closeFileFn()
				return gctx.Err()
			}
		}

		return nil
	})

	g.Go(func() error {
		opts := minio.SnowballOptions{
			InMemory: false,
			Compress: true,
			SkipErrs: false,
		}

		//	Put resource into bucket using tar archive containing all uploaded resources
		if err := s.client.PutObjectsSnowball(ctx, Bucket, opts, objsCh); err != nil {
			s.log.Error("minio: failed to put objects into bucket", zap.Error(err))
			return err
		}

		return nil
	})

	//	Wait for MinIO to finish upload
	if err := g.Wait(); err != nil {
		return nil, err
	}

	//	Commit TX
	if err := tx.Commit(ctx); err != nil {
		s.log.Error("tx: failed to commit", zap.Error(err))
		return nil, err
	}

	//	Convert models into response dto
	resp := make(response.ResourceInfoList, 0, len(uploadResources))
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
		if !errors.Is(err, errs.ErrResourceNotFound) {
			s.log.Error("resource repo: failed to get resource", zap.Error(err))
		}

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
		s.log.Error("failed to search resource", zap.Error(err))
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

	content := make([]models.Resource, 0)

	if resource.IsDir() {
		//	Get dir content
		content, err = s.dirRepo.GetAll(ctx, resource, true)
		if err != nil {
			return err
		}
	}

	//	Delete from Bucket
	if resource.IsDir() && len(content) > 0 {
		objsCh := make(chan minio.ObjectInfo)

		go func() {
			defer close(objsCh)

			for _, c := range content {
				objsCh <- minio.ObjectInfo{
					Key: c.ObjectKey(),
				}
			}
		}()

		//	Delete directory content first
		minioErr := s.client.RemoveObjects(ctx, Bucket, objsCh, minio.RemoveObjectsOptions{})
		for rmErr := range minioErr {
			if rmErr.Err != nil {
				return rmErr.Err
			}
		}
	} else {
		if err = s.client.RemoveObject(
			ctx,
			Bucket,
			resource.ObjectKey(),
			minio.RemoveObjectOptions{},
		); err != nil {
			return err
		}
	}

	//	Delete from DB
	if err = s.resourceRepo.Delete(ctx, &resource); err != nil {
		if !errors.Is(err, errs.ErrResourceNotFound) {
			s.log.Error("resource repo: failed to get resource", zap.Error(err))
		}

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
		if !errors.Is(err, errs.ErrResourceNotFound) {
			s.log.Error("resource repo: failed to get resource", zap.Error(err))
		}

		return DownloadResult{}, err
	}

	//	Get actual resource from MinIO
	content, err := s.client.GetObject(ctx, Bucket, resource.ObjectKey(), minio.GetObjectOptions{})
	if err != nil {
		return DownloadResult{}, err
	}

	info, err := content.Stat()
	if err != nil {
		return DownloadResult{}, err
	}

	return DownloadResult{
		Content: content,
		Name:    resource.Name,
		Size:    info.Size,
		Type:    info.ContentType,
	}, nil
}
