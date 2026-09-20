package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	corectx "github.com/Nurlan270/cloud-storage-go/internal/core/context"
	errs "github.com/Nurlan270/cloud-storage-go/internal/core/errors"
	"github.com/Nurlan270/cloud-storage-go/internal/core/models"
)

type DirectoryRepository interface {
	GetAll(ctx context.Context, userID uint64, path string) ([]models.Resource, error)
	Create(ctx context.Context, dir models.Resource) (models.Resource, error)
	Exists(ctx context.Context, dir models.Resource) (bool, error)
}

type directoryRepository struct {
	pool *pgxpool.Pool
}

func NewDirectoryRepository(pool *pgxpool.Pool) DirectoryRepository {
	return &directoryRepository{pool: pool}
}

func (r *directoryRepository) GetAll(
	ctx context.Context,
	userID uint64,
	path string,
) ([]models.Resource, error) {
	const q = `
		SELECT path, name, size, type
		FROM resources
		WHERE user_id = $1 AND path = $2
	`

	rows, err := r.pool.Query(ctx, q, userID, path)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrDirectoryNotFound
		}

		return nil, err
	}

	var resources []models.Resource

	for rows.Next() {
		var resource models.Resource

		if err = rows.Scan(&resource.Path, &resource.Name, &resource.Size, &resource.Type); err != nil {
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

	var exists bool
	if err := r.pool.QueryRow(
		ctx, q,
		dir.UserID, dir.Path, dir.Name, dir.Type,
	).Scan(&exists); err != nil {
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
