package repository

import (
	"database/sql"
	"errors"

	errs "github.com/Nurlan270/cloud-storage-go/internal/core/errors"
	"github.com/Nurlan270/cloud-storage-go/internal/core/models"
)

type UserRepository interface {
	CreateUser(user *models.User) (*models.User, error)
	GetUserFromUserID(userID uint64) (*models.User, error)
	GetUserFromUsername(username string) (*models.User, error)
}

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return userRepository{db: db}
}

func (r userRepository) CreateUser(user *models.User) (*models.User, error) {
	const q = "INSERT INTO users (username, password) VALUES ($1, $2) RETURNING id, username"

	u := &models.User{}
	if err := r.db.QueryRow(q, user.Username, user.Password).Scan(&u.ID, &u.Username); err != nil {
		return nil, err
	}

	return u, nil
}

func (r userRepository) GetUserFromUsername(username string) (*models.User, error) {
	const q = "SELECT id, username, password FROM users WHERE username = $1"

	u := &models.User{}
	if err := r.db.QueryRow(q, username).Scan(&u.ID, &u.Username, &u.Password); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.ErrUserNotFound
		}

		return nil, err
	}

	return u, nil
}

func (r userRepository) GetUserFromUserID(userID uint64) (*models.User, error) {
	const q = "SELECT id, username, password FROM users WHERE id = $1"

	u := &models.User{}
	if err := r.db.QueryRow(q, userID).Scan(&u.ID, &u.Username, &u.Password); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.ErrUserNotFound
		}

		return nil, err
	}

	return u, nil
}
