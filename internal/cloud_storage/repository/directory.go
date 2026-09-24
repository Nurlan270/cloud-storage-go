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
}

type directoryRepository struct {
	pool *pgxpool.Pool

	log *zap.Logger
}

func NewDirectoryRepository(pool *pgxpool.Pool) DirectoryRepository {
	log := logger.Get().With(
		zap.String("src", "directory repository"))

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

	resources := make([]models.Resource, 0)

	for rows.Next() {
		var resource models.Resource

		if err = rows.Scan(
			&resource.UserID,
			&resource.Path,
			&resource.Name,
			&resource.Size,
			&resource.Type,
		); err != nil {
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
		INSERT INTO resources (user_id, path, name, type)
		VALUES ($1, $2, $3, $4)
		RETURNING user_id, path, name, type
	`

	tx := corectx.TxFromContext(ctx)

	res := models.Resource{}
	if err := tx.QueryRow(
		ctx, q,
		dir.UserID, dir.Path, dir.Name, models.TypeDir,
	).Scan(
		&res.UserID, &res.Path, &res.Name, &res.Type,
	); err != nil {
		if errIs(err, pgerrcode.UniqueViolation) {
			return res, errs.ErrDirectoryAlreadyExists
		} else {
			r.log.Error("failed to create", zap.Error(err))
		}

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

func errIs(err error, errCode string) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == errCode {
		return true
	}

	return false
}
