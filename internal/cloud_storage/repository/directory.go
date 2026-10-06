package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	corectx "github.com/Nurlan270/cloud-storage-go/internal/core/context"
	errs "github.com/Nurlan270/cloud-storage-go/internal/core/errors"
	"github.com/Nurlan270/cloud-storage-go/internal/core/logger"
	"github.com/Nurlan270/cloud-storage-go/internal/core/models"
)

type DirectoryRepository interface {
	GetAll(ctx context.Context, dir models.Resource, recursive bool) ([]models.Resource, error)
	Create(ctx context.Context, dir models.Resource) (models.Resource, error)
	Exists(ctx context.Context, dir models.Resource) (bool, error)
	Delete(ctx context.Context, dir models.Resource) error
	BatchUpdate(
		ctx context.Context,
		old models.Resource,
		new models.Resource,
		content []models.Resource,
	) error
}

type directoryRepository struct {
	pool *pgxpool.Pool

	log *zap.Logger
}

func NewDirectoryRepository(pool *pgxpool.Pool) DirectoryRepository {
	log := logger.Get().SetSrc("directory repository")

	return &directoryRepository{
		pool: pool,
		log:  log,
	}
}

func (r *directoryRepository) GetAll(
	ctx context.Context,
	dir models.Resource,
	recursive bool,
) ([]models.Resource, error) {
	const (
		qLinear = `
			SELECT user_id, path, name, size, type
			FROM resources
			WHERE user_id = $1 AND path = $2
		`

		qRecursive = `
			SELECT user_id, path, name, size, type
			FROM resources
			WHERE user_id = $1 AND path LIKE $2
		`
	)

	// Check whether provided dir exists
	if exists, err := r.Exists(ctx, dir); err != nil {
		return nil, err
	} else if !exists {
		return nil, errs.ErrDirectoryNotFound
	}

	var (
		rows pgx.Rows
		err  error
	)

	if recursive {
		//	Get dir content recursively
		rows, err = r.pool.Query(ctx, qRecursive, dir.UserID, dir.FullPath()+"%")
		if err != nil {
			r.log.Error("failed to get all recursively", zap.Error(err))

			return nil, err
		}
	} else {
		//	Get dir content linearly
		rows, err = r.pool.Query(ctx, qLinear, dir.UserID, dir.FullPath())
		if err != nil {
			r.log.Error("failed to get all linearly", zap.Error(err))

			return nil, err
		}
	}

	resources := make([]models.Resource, 0, rows.CommandTag().RowsAffected())

	for rows.Next() {
		var resource models.Resource

		if err = rows.Scan(
			&resource.UserID, &resource.Path, &resource.Name, &resource.Size, &resource.Type,
		); err != nil {
			r.log.Error("failed to scan row", zap.Error(err))

			return nil, err
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

func (r *directoryRepository) Create(
	ctx context.Context,
	dir models.Resource,
) (models.Resource, error) {
	const q = `
		INSERT INTO resources (user_id, path, name, size, type)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING user_id, path, name, type
	`

	tx := corectx.TxFromContext(ctx)

	res := models.Resource{}
	if err := tx.QueryRow(
		ctx, q,
		dir.UserID, dir.Path, dir.Name, 0, models.TypeDir,
	).Scan(
		&res.UserID, &res.Path, &res.Name, &res.Type,
	); err != nil {
		if errIs(err, pgerrcode.UniqueViolation) {
			return res, errs.ErrDirectoryAlreadyExists
		}

		r.log.Error("failed to create", zap.Error(err))

		return res, err
	}

	return res, nil
}

func (r *directoryRepository) Exists(ctx context.Context, dir models.Resource) (bool, error) {
	const q = `
		SELECT EXISTS(
			SELECT 1 FROM resources
			WHERE user_id = $1 AND path = $2 AND name = $3 AND type = $4
		)
	`

	if dir.Path == "/" && dir.Name == "" {
		return true, nil
	}

	var exists bool
	if err := r.pool.QueryRow(
		ctx, q,
		dir.UserID, dir.Path, dir.Name, dir.Type,
	).Scan(&exists); err != nil {
		r.log.Error("failed to check for existence", zap.Error(err))

		return exists, err
	}

	return exists, nil
}

func (r *directoryRepository) Delete(ctx context.Context, dir models.Resource) error {
	const (
		qDeleteContent = `
			DELETE FROM resources
			WHERE user_id = $1 AND path LIKE $2
		`

		qDeleteDir = `
			DELETE FROM resources
			WHERE user_id = $1 AND type = $2 AND path = $3 AND name = $4
		`
	)

	tx := corectx.TxFromContext(ctx)

	//	Remove directory content
	q1, err := tx.Exec(ctx, qDeleteContent, dir.UserID, dir.FullPath()+"%")
	if err != nil {
		r.log.Error("failed to delete dir content", zap.Error(err))

		return err
	}

	//	Remove directory itself
	q2, err := tx.Exec(ctx, qDeleteDir, dir.UserID, dir.Type, dir.Path, dir.Name)
	if err != nil {
		r.log.Error("failed to delete dir", zap.Error(err))

		return err
	}

	if (q1.RowsAffected() + q2.RowsAffected()) <= 0 {
		return errs.ErrDirectoryNotFound
	}

	return nil
}

func (r *directoryRepository) BatchUpdate(
	ctx context.Context,
	old models.Resource,
	new models.Resource,
	content []models.Resource,
) error {
	const (
		qUpdateOldResourceContent = `
			UPDATE resources
			SET path = $1
			WHERE user_id = $2 AND type = $3 AND path = $4 AND name = $5
			RETURNING user_id, path, name, size, type
		`

		qUpdateOldResource = `
			UPDATE resources
			SET name = $1, path = $2
			WHERE user_id = $3 AND type = $4 AND path = $5 AND name = $6
		`
	)

	tx := corectx.TxFromContext(ctx)

	batch := &pgx.Batch{}

	//	Update old resource with new one
	batch.Queue(qUpdateOldResource,
		new.Name, new.Path,
		old.UserID, old.Type, old.Path, old.Name,
	)

	//	Update old resource's content with new one
	for _, oldContent := range content {
		batch.Queue(qUpdateOldResourceContent,
			new.FullPath(), // new content's path
			oldContent.UserID, oldContent.Type, oldContent.Path, oldContent.Name,
		)
	}

	br := tx.SendBatch(ctx, batch)
	defer br.Close()

	for range batch.QueuedQueries {
		if _, err := br.Exec(); err != nil {
			if errIs(err, pgerrcode.UniqueViolation) {
				return errs.ErrResourceAlreadyExists
			}

			r.log.Error("failed to batch update", zap.Error(err))

			return err
		}
	}

	return nil
}

func errIs(err error, errCode string) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == errCode {
		return true
	}

	return false
}
