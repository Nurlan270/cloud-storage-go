package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	errs "github.com/Nurlan270/cloud-storage-go/internal/core/errors"
	"github.com/Nurlan270/cloud-storage-go/internal/core/models"
)

type UserRepository interface {
	CreateUser(user *models.User) (*models.User, error)
	GetUserFromUserID(userID uint64) (*models.User, error)
	GetUserFromUsername(username string) (*models.User, error)
}

type userRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) UserRepository {
	return userRepository{pool: pool}
}

func (r userRepository) CreateUser(user *models.User) (*models.User, error) {
	const q = "INSERT INTO users (username, password) VALUES ($1, $2) RETURNING id, username"

	ctx := context.Background()

	u := &models.User{}
	if err := r.pool.QueryRow(ctx, q, user.Username, user.Password).Scan(&u.ID, &u.Username); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return nil, errs.ErrUserAlreadyExists
		}

		return nil, err
	}

	return u, nil
}

func (r userRepository) GetUserFromUsername(username string) (*models.User, error) {
	const q = "SELECT id, username, password FROM users WHERE username = $1"

	ctx := context.Background()

	u := &models.User{}
	if err := r.pool.QueryRow(ctx, q, username).Scan(&u.ID, &u.Username, &u.Password); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrUserNotFound
		}

		return nil, err
	}

	return u, nil
}

func (r userRepository) GetUserFromUserID(userID uint64) (*models.User, error) {
	const q = "SELECT id, username, password FROM users WHERE id = $1"

	ctx := context.Background()

	u := &models.User{}
	if err := r.pool.QueryRow(ctx, q, userID).Scan(&u.ID, &u.Username, &u.Password); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errs.ErrUserNotFound
		}

		return nil, err
	}

	return u, nil
}
