package repository

import (
	"database/sql"

	"github.com/Nurlan270/cloud-storage-go/internal/core/models"
)

type UserRepository interface {
	CreateUser(username, password string) (*models.User, error)
}

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return userRepository{db: db}
}

func (r userRepository) CreateUser(username, password string) (*models.User, error) {
	const q = "INSERT INTO users (username, password) VALUES ($1, $2) RETURNING id, username"

	u := &models.User{}
	if err := r.db.QueryRow(q, username, password).Scan(&u.ID, &u.Username); err != nil {
		return nil, err
	}

	return u, nil
}
