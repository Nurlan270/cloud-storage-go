package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	errs "github.com/Nurlan270/cloud-storage-go/internal/core/errors"
	"github.com/Nurlan270/cloud-storage-go/internal/core/models"
)

type ResourceRepository interface {
	BatchInsertTx(ctx context.Context, tx pgx.Tx, resources []models.Resource) error
	Get(
		ctx context.Context,
		userID uint64,
		path, name string,
		resourceType models.ResourceType,
	) (*models.Resource, error)
	Search(ctx context.Context, userID uint64, query string) ([]*models.Resource, error)
	DeleteTx(
		ctx context.Context,
		tx pgx.Tx,
		userID uint64,
		path, name string,
		resourceType models.ResourceType,
	) error
}

type resourceRepository struct {
	pool *pgxpool.Pool
}

func NewResourceRepository(pool *pgxpool.Pool) ResourceRepository {
	return &resourceRepository{pool: pool}
}

func (r *resourceRepository) BatchInsertTx(
	ctx context.Context,
	tx pgx.Tx,
	resources []models.Resource,
) error {
	const q = `
		INSERT INTO resources (user_id, path, name, size, type)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user_id, path, name, type) DO NOTHING
	`

	batch := &pgx.Batch{}
	for _, res := range resources {
		batch.Queue(q, res.UserID, res.Path, res.Name, res.Size, res.Type)
	}

	br := tx.SendBatch(ctx, batch)
	defer br.Close()

	var inserted int64

	for range resources {
		tag, err := br.Exec()
		if err != nil {
			return err
		}

		inserted += tag.RowsAffected()
	}

	if inserted == 0 && len(resources) > 0 {
		//	Return error only if no rows were affected
		//	if at least 1 row was inserted return nil
		return errs.ErrResourceAlreadyExists
	}

	return nil
}

func (r *resourceRepository) Get(
	ctx context.Context,
	userID uint64,
	path, name string,
	resourceType models.ResourceType,
) (*models.Resource, error) {
	const q = `
		SELECT user_id, path, name, size, type
		FROM resources
		WHERE user_id = $1 AND type = $2 AND path = $3 AND name = $4
	`

	res := &models.Resource{}
	if err := r.pool.QueryRow(
		ctx, q,
		userID,
		resourceType,
		path,
		name,
	).Scan(
		&res.UserID,
		&res.Path,
		&res.Name,
		&res.Size,
		&res.Type,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.ErrResourceNotFound
		}

		return nil, err
	}

	return res, nil
}

func (r *resourceRepository) Search(
	ctx context.Context,
	userID uint64,
	query string,
) ([]*models.Resource, error) {
	const q = `
		SELECT user_id, path, name, size, type
		FROM resources
		WHERE user_id = $1 AND (path ILIKE $2 OR name ILIKE $2)
	`

	rows, err := r.pool.Query(ctx, q, userID, "%"+query+"%")
	if err != nil {
		return nil, err
	}

	var list []*models.Resource

	for rows.Next() {
		res := &models.Resource{}

		if err = rows.Scan(&res.UserID, &res.Path, &res.Name, &res.Size, &res.Type); err != nil {
			return nil, err
		}

		list = append(list, res)
	}

	return list, nil
}

func (r *resourceRepository) DeleteTx(
	ctx context.Context,
	tx pgx.Tx,
	userID uint64,
	path, name string,
	resourceType models.ResourceType,
) error {
	const (
		qSingleDelete = `
			DELETE FROM resources
			WHERE user_id = $1 AND type = $2 AND path = $3 AND name = $4
		`

		qDeleteAll = `
			DELETE FROM resources
			WHERE user_id = $1 AND path LIKE $2
		`
	)

	var deleted int64

	if resourceType == models.TypeDir {
		//	Remove provided folder
		tag, err := tx.Exec(ctx, qSingleDelete, userID, resourceType, path, name)
		if err != nil {
			return err
		}

		deleted += tag.RowsAffected()

		//	Remove all resources that's within provided folder
		dirPath := strings.TrimLeft(path+name, "/") + "/"

		tag, err = tx.Exec(ctx, qDeleteAll, userID, dirPath+"%")
		if err != nil {
			return err
		}

		deleted += tag.RowsAffected()
	} else {
		tag, err := tx.Exec(ctx, qSingleDelete, userID, resourceType, path, name)
		if err != nil {
			return err
		}

		deleted += tag.RowsAffected()
	}

	if deleted <= 0 {
		return errs.ErrResourceNotFound
	}

	return nil
}
