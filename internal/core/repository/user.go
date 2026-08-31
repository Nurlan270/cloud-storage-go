package repository

import (
	"database/sql"

	"github.com/Nurlan270/cloud-storage-go/internal/core/models"
)

type UserRepository interface {
	CreateUser(user *models.User) (*models.User, error)
	GetUser(user *models.User) (*models.User, error)
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

func (r userRepository) GetUser(user *models.User) (*models.User, error) {
	const q = "SELECT id, username, password FROM users WHERE username = $1"

	u := &models.User{}
	if err := r.db.QueryRow(q, user.Username).Scan(&u.ID, &u.Username, &u.Password); err != nil {
		return nil, err
	}

	return u, nil
}
