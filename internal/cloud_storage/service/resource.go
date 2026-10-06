package service

import (
	"context"
	"io"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/Nurlan270/cloud-storage-go/internal/cloud_storage/transport/http/dto/request"
	"github.com/Nurlan270/cloud-storage-go/internal/core/archiver"
	corectx "github.com/Nurlan270/cloud-storage-go/internal/core/context"
	errs "github.com/Nurlan270/cloud-storage-go/internal/core/errors"
	"github.com/Nurlan270/cloud-storage-go/internal/core/logger"
	"github.com/Nurlan270/cloud-storage-go/internal/core/minio"
	"github.com/Nurlan270/cloud-storage-go/internal/core/models"
)

type ResourceService interface {
	Upload(ctx context.Context, req request.UploadResource) ([]models.Resource, error)
	GetInfo(ctx context.Context, req request.GetResourceInfo) (models.Resource, error)
	Search(ctx context.Context, req request.SearchResource) ([]models.Resource, error)
	Move(ctx context.Context, req request.MoveResource) (models.Resource, error)
	Download(ctx context.Context, req request.DownloadResource) (DownloadResult, error)
	Delete(ctx context.Context, req request.DeleteResource) error
}

type ResourceRepository interface {
	BatchCreate(ctx context.Context, resources []models.Resource) error
	Get(ctx context.Context, resource models.Resource) (models.Resource, error)
	Update(ctx context.Context, old models.Resource, new models.Resource) (models.Resource, error)
	Search(ctx context.Context, userID uint64, query string) ([]models.Resource, error)
	Delete(ctx context.Context, resource models.Resource) error
}

type MinioClient interface {
	Get(ctx context.Context, resource models.Resource) (minio.GetResult, error)
	GetAll(ctx context.Context, resources []models.Resource) ([]minio.GetResult, error)
	Put(ctx context.Context, entity minio.PutEntity) error
	PutAll(ctx context.Context, entities []minio.PutEntity) error
	Update(ctx context.Context, old models.Resource, new models.Resource) error
	UpdateAll(
		ctx context.Context,
		old models.Resource,
		new models.Resource,
		resources []models.Resource,
	) error
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
	log := logger.Get().SetSrc("resource service")

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
) ([]models.Resource, error) {
	user := corectx.UserFromContext(ctx)

	resourceSet := make(map[models.Resource]struct{}, len(req.Objects))
	resourceList := make([]models.Resource, 0, len(req.Objects)*2)
	uploadEntities := make([]minio.PutEntity, 0, len(req.Objects))

	for _, obj := range req.Objects {
		fullPath, err := getFullPath(obj.Header, req.Path)
		if err != nil {
			return nil, err
		}

		resources := buildResourcesFromPath(user.ID, fullPath, obj.Size)

		for _, resource := range resources {
			//	Drop duplicate values
			if _, exists := resourceSet[resource]; !exists {
				resourceSet[resource] = struct{}{}
				resourceList = append(resourceList, resource)
			}

			if !resource.IsDir() {
				uploadEntities = append(uploadEntities, minio.PutEntity{
					Resource: resource,
					Object:   obj,
				})
			}
		}
	}

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
		return nil, err
	}

	//	Commit TX
	if err := tx.Commit(ctx); err != nil {
		s.log.Error("tx: failed to commit", zap.Error(err))
		return nil, err
	}

	return resourceList, nil
}

func (s *resourceService) GetInfo(
	ctx context.Context,
	req request.GetResourceInfo,
) (models.Resource, error) {
	user := corectx.UserFromContext(ctx)

	path, name := splitPath(req.Path)
	resourceType := getResourceType(req.Path)
	resource := models.Resource{
		UserID: user.ID,
		Path:   path,
		Name:   name,
		Type:   resourceType,
	}

	res, err := s.resourceRepo.Get(ctx, resource)
	if err != nil {
		return models.Resource{}, err
	}

	return res, nil
}

func (s *resourceService) Search(
	ctx context.Context,
	req request.SearchResource,
) ([]models.Resource, error) {
	user := corectx.UserFromContext(ctx)

	list, err := s.resourceRepo.Search(ctx, user.ID, req.Query)
	if err != nil {
		return nil, err
	}

	return list, nil
}

func (s *resourceService) Delete(ctx context.Context, req request.DeleteResource) error {
	user := corectx.UserFromContext(ctx)
	path, name := splitPath(req.Path)
	resourceType := getResourceType(req.Path)
	resource := models.Resource{
		UserID: user.ID,
		Path:   path,
		Name:   name,
		Type:   resourceType,
	}

	//	Start TX
	tx, txErr := s.pool.Begin(ctx)
	if txErr != nil {
		s.log.Error("tx: failed to start", zap.Error(txErr))
		return txErr
	}
	defer tx.Rollback(ctx)

	//	Put TX into ctx
	ctx = corectx.NewTxContext(ctx, tx)

	if resource.IsDir() {
		//	Get directory content
		content, err := s.dirRepo.GetAll(ctx, resource, true)
		if err != nil {
			return err
		}

		//	Delete from DB
		if err = s.dirRepo.Delete(ctx, resource); err != nil {
			return err
		}

		//	Delete dir content from bucket
		if err = s.client.DeleteAll(ctx, content); err != nil {
			return err
		}

		//	Delete dir from bucket
		if err = s.client.Delete(ctx, resource); err != nil {
			return err
		}
	} else {
		//	Delete from DB
		if err := s.resourceRepo.Delete(ctx, resource); err != nil {
			return err
		}

		//	Delete from bucket
		if err := s.client.Delete(ctx, resource); err != nil {
			return err
		}
	}

	//	Commit TX
	if err := tx.Commit(ctx); err != nil {
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

	//	Function to remove downloaded archive file after streaming to user
	Remove func()
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

	var result DownloadResult

	//	Check whether provided resource exists
	if _, err := s.resourceRepo.Get(ctx, resource); err != nil {
		return result, err
	}

	//	Get actual resource from MinIO
	if resource.IsDir() {
		//	Get directory content
		content, err := s.dirRepo.GetAll(ctx, resource, true)
		if err != nil {
			return result, err
		}

		//	Get resources from bucket
		resources, err := s.client.GetAll(ctx, content)
		if err != nil {
			return result, err
		}

		archiveFiles := make([]archiver.ArchiveFile, 0, len(resources))

		for _, res := range resources {
			//	Create tmp file
			file, err := os.CreateTemp(archiver.ArchiveFilesDir, res.Name)
			if err != nil {
				s.log.Error("failed to create tmp file", zap.Error(err))

				s.closeFile(res.Content)

				return result, err
			}

			//	Copy content into tmp file
			if _, err = io.Copy(file, res.Content); err != nil {
				s.log.Error("failed to copy object content into tmp file", zap.Error(err))

				s.removeFile(file.Name())
				s.closeFile(res.Content)

				return result, err
			}

			s.closeFile(res.Content)

			archiveFiles = append(archiveFiles, archiver.ArchiveFile{
				PathOnDisk: file.Name(),
				Filename:   res.FullPath,
			})
		}

		archiveResult, err := archiver.Archive(ctx, resource.Name, archiveFiles)
		if err != nil {
			return result, err
		}

		result = DownloadResult{
			Content: archiveResult.Content,
			Name:    archiveResult.Name,
			Size:    archiveResult.Size,
			Type:    "application/zip",
			Remove: func() {
				if err = os.Remove(archiveResult.Path); err != nil {
					s.log.Warn("failed to remove archive file", zap.Error(err))
				}
			},
		}
	} else {
		getResult, err := s.client.Get(ctx, resource)
		if err != nil {
			return result, err
		}

		result = DownloadResult{
			Content: getResult.Content,
			Name:    getResult.Name,
			Size:    getResult.Size,
			Type:    getResult.ContentType,
		}
	}

	return result, nil
}

func (s *resourceService) Move(ctx context.Context, req request.MoveResource) (models.Resource, error) {
	user := corectx.UserFromContext(ctx)

	//	Old resource data
	oldPath, oldName := splitPath(req.From)
	oldResourceType := getResourceType(req.From)

	oldResource := models.Resource{
		UserID: user.ID,
		Path:   oldPath,
		Name:   oldName,
		Type:   oldResourceType,
	}

	//	New resource data
	newPath, newName := splitPath(req.To)
	newResourceType := getResourceType(req.To)

	newResource := models.Resource{
		UserID: user.ID,
		Path:   newPath,
		Name:   newName,
		Type:   newResourceType,
	}

	//	Check whether old & new resources type is identical
	if oldResource.Type != newResource.Type {
		return models.Resource{}, errs.ErrResourceNonIdenticalTypes
	}

	//	Start TX
	tx, txErr := s.pool.Begin(ctx)
	if txErr != nil {
		s.log.Error("tx: failed to start", zap.Error(txErr))
		return models.Resource{}, txErr
	}
	defer tx.Rollback(ctx)

	//	Put TX into ctx
	ctx = corectx.NewTxContext(ctx, tx)

	if newResource.IsDir() {
		//	Get directory content
		content, err := s.dirRepo.GetAll(ctx, oldResource, true)
		if err != nil {
			return models.Resource{}, err
		}

		//	Update all directory content in DB
		if err = s.dirRepo.BatchUpdate(ctx, oldResource, newResource, content); err != nil {
			return models.Resource{}, err
		}

		//	Update all directory content in bucket
		if err = s.client.UpdateAll(ctx, oldResource, newResource, content); err != nil {
			return models.Resource{}, err
		}
	} else {
		//	Update resource in DB
		_, err := s.resourceRepo.Update(ctx, oldResource, newResource)
		if err != nil {
			return models.Resource{}, err
		}

		//	Update resource in bucket
		if err = s.client.Update(ctx, oldResource, newResource); err != nil {
			return models.Resource{}, err
		}
	}

	//	Commit TX
	if err := tx.Commit(ctx); err != nil {
		s.log.Error("tx: failed to commit", zap.Error(err))
		return models.Resource{}, err
	}

	return models.Resource{
		Path: newResource.Path,
		Name: newResource.Name,
		Size: newResource.Size,
		Type: newResource.Type,
	}, nil
}

func (s *resourceService) closeFile(file io.Closer) {
	if err := file.Close(); err != nil {
		s.log.Warn("failed to close file", zap.Error(err))
	}
}

func (s *resourceService) removeFile(filename string) {
	if err := os.Remove(filename); err != nil {
		s.log.Warn("failed to remove file",
			zap.String("filename", filename), zap.Error(err))
	}
}
