package repository

import (
	"context"
	"errors"
	corectx "github.com/Nurlan270/cloud-storage-go/internal/core/context"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	errs "github.com/Nurlan270/cloud-storage-go/internal/core/errors"
	"github.com/Nurlan270/cloud-storage-go/internal/core/models"
)

type ResourceRepository interface {
	BatchInsert(ctx context.Context, resources []models.Resource) error
	Get(ctx context.Context, resource *models.Resource) (*models.Resource, error)
	Update(ctx context.Context, old *models.Resource, new *models.Resource) (*models.Resource, error)
	Search(ctx context.Context, userID uint64, query string) ([]*models.Resource, error)
	Delete(ctx context.Context, resource *models.Resource) error
}

type resourceRepository struct {
	pool *pgxpool.Pool
}

func NewResourceRepository(pool *pgxpool.Pool) ResourceRepository {
	return &resourceRepository{pool: pool}
}

func (r *resourceRepository) BatchInsert(
	ctx context.Context,
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

	tx := corectx.TxFromContext(ctx)

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

func (r *resourceRepository) Get(ctx context.Context, resource *models.Resource) (*models.Resource, error) {
	const q = `
		SELECT user_id, path, name, size, type
		FROM resources
		WHERE user_id = $1 AND type = $2 AND path = $3 AND name = $4
	`

	res := &models.Resource{}
	if err := r.pool.QueryRow(
		ctx, q,
		resource.UserID, resource.Type, resource.Path, resource.Name,
	).Scan(
		&res.UserID, &res.Path, &res.Name, &res.Size, &res.Type,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
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

func (r *resourceRepository) Delete(ctx context.Context, resource *models.Resource) error {
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

	tx := corectx.TxFromContext(ctx)

	var deleted int64

	if resource.IsDir() {
		//	Remove provided folder
		tag, err := tx.Exec(ctx, qSingleDelete, resource.UserID, resource.Type, resource.Path, resource.Name)
		if err != nil {
			return err
		}

		deleted += tag.RowsAffected()

		dirPath := strings.TrimLeft(resource.FullPath(), "/") + "/"

		//	Remove all resources that's within provided folder
		tag, err = tx.Exec(ctx, qDeleteAll, resource.UserID, dirPath+"%")
		if err != nil {
			return err
		}

		deleted += tag.RowsAffected()
	} else {
		tag, err := tx.Exec(ctx, qSingleDelete, resource.UserID, resource.Type, resource.Path, resource.Name)
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

func (r *resourceRepository) Update(ctx context.Context, old *models.Resource, new *models.Resource) (*models.Resource, error) {
	const q = `
		UPDATE resources
		SET name = $1, path = $2, type = $3
		WHERE user_id = $4 AND type = $5 AND path = $6 AND name = $7
		RETURNING user_id, path, name, size, type
	`

	tx := corectx.TxFromContext(ctx)

	res := &models.Resource{}
	if err := tx.QueryRow(
		ctx, q,
		new.Name, new.Path, new.Type,
		old.UserID, old.Type, old.Path, old.Name,
	).Scan(
		&res.UserID, &res.Path, &res.Name, &res.Size, &res.Type,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrResourceNotFound
		}

		return nil, err
	}

	return res, nil
}
