package minio

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"

	"github.com/minio/minio-go/v7"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/Nurlan270/cloud-storage-go/internal/core/logger"
	"github.com/Nurlan270/cloud-storage-go/internal/core/models"
)

type Client interface {
	Get(ctx context.Context, resource models.Resource) (GetResult, error)
	Put(ctx context.Context, opts PutOptions) error
	PutAll(ctx context.Context, entities []PutAllEntities) error
	Delete(ctx context.Context, resource models.Resource) error
	DeleteAll(ctx context.Context, resources []models.Resource) error
}

type client struct {
	*minio.Client

	log *zap.Logger
}

func MustNew(conf Config) Client {
	c, err := connect(conf)
	if err != nil {
		panic(fmt.Sprintf("minio client: %v", err))
	}

	log := logger.Get().With(
		zap.String("src", "minio client"))

	return &client{
		Client: c,
		log:    log,
	}
}

// Bucket is the name of main bucket.
const Bucket = "user-files"

type GetResult struct {
	Content     io.ReadCloser
	Name        string
	Size        int64
	ContentType string
}

func (c *client) Get(ctx context.Context, resource models.Resource) (GetResult, error) {
	content, err := c.GetObject(ctx, Bucket, resource.ObjectKey(), minio.GetObjectOptions{})
	if err != nil {
		c.log.Error("failed to get object",
			zap.Any("object", resource), zap.Error(err))

		return GetResult{}, err
	}

	info, err := content.Stat()
	if err != nil {
		c.log.Error("failed to stat object",
			zap.Any("object", resource), zap.Error(err))

		return GetResult{}, err
	}

	return GetResult{
		Content:     content,
		Name:        resource.Name,
		Size:        info.Size,
		ContentType: info.ContentType,
	}, nil
}

type PutOptions struct {
	Reader      io.Reader
	Key         string
	Size        int64
	ContentType string
}

func (c *client) Put(ctx context.Context, opts PutOptions) error {
	minioOpts := minio.PutObjectOptions{
		ContentType: opts.ContentType,
	}

	if _, err := c.PutObject(
		ctx, Bucket,
		opts.Key, opts.Reader, opts.Size,
		minioOpts,
	); err != nil {
		c.log.Error("failed to Put", zap.Error(err))

		return err
	}

	return nil
}

type PutAllEntities struct {
	Resource models.Resource
	Object   *multipart.FileHeader
}

func (c *client) PutAll(ctx context.Context, entities []PutAllEntities) error {
	objsCh := make(chan minio.SnowballObject)

	g, gctx := errgroup.WithContext(ctx)

	//	Producer
	g.Go(func() error {
		defer close(objsCh)

		for _, entity := range entities {
			file, err := entity.Object.Open()
			if err != nil {
				return fmt.Errorf("failed to open file: %w", err)
			}

			//	File closer
			closeFileFn := func() {
				if fileErr := file.Close(); fileErr != nil {
					c.log.Warn("failed to close file",
						zap.Any("file", file), zap.Error(fileErr),
					)
				}
			}

			obj := minio.SnowballObject{
				Key:     entity.Resource.ObjectKey(),
				Size:    *entity.Resource.Size,
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

	//	Consumer
	g.Go(func() error {
		minioOpts := minio.SnowballOptions{
			InMemory: false,
			Compress: true,
			SkipErrs: false,
		}

		if err := c.PutObjectsSnowball(gctx, Bucket, minioOpts, objsCh); err != nil {
			return fmt.Errorf("failed to put objects: %w", err)
		}

		return nil
	})

	if err := g.Wait(); err != nil {
		c.log.Error("failed to PutAll", zap.Error(err))
		return err
	}

	return nil
}

func (c *client) Delete(ctx context.Context, resource models.Resource) error {
	opts := minio.RemoveObjectOptions{
		ForceDelete: true,
	}

	if err := c.RemoveObject(
		ctx, Bucket,
		resource.ObjectKey(), opts,
	); err != nil {
		c.log.Error("failed to Delete",
			zap.Any("object", resource), zap.Error(err))

		return err
	}

	return nil
}

func (c *client) DeleteAll(ctx context.Context, resources []models.Resource) error {
	objsCh := make(chan minio.ObjectInfo)

	go func() {
		defer close(objsCh)

		for _, r := range resources {
			select {
			case objsCh <- minio.ObjectInfo{
				Key: r.ObjectKey(),
			}:
			case <-ctx.Done():
				return
			}
		}
	}()

	errCh := c.RemoveObjects(ctx, Bucket, objsCh, minio.RemoveObjectsOptions{})
	for err := range errCh {
		if err.Err != nil {
			c.log.Error("failed to DeleteAll", zap.Error(err.Err))

			return err.Err
		}
	}

	return ctx.Err()
}
