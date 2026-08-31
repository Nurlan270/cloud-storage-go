package repository

import (
	"database/sql"

	"github.com/Nurlan270/cloud-storage-go/internal/core/models"
)

type SessionRepository interface {
	CreateSession(session *models.Session) (*models.Session, error)
}

type sessionRepository struct {
	db *sql.DB
}

func NewSessionRepository(db *sql.DB) SessionRepository {
	return sessionRepository{db: db}
}

func (r sessionRepository) CreateSession(session *models.Session) (*models.Session, error) {
	const q = `
		INSERT INTO sessions (uuid, user_id, expires_at)
		VALUES ($1, $2, $3)
		RETURNING uuid, user_id, expires_at
	`

	s := &models.Session{}
	if err := r.db.QueryRow(
		q,
		session.UUID,
		session.UserID,
		session.ExpiresAt,
	).Scan(
		&s.UUID,
		&s.UserID,
		&s.ExpiresAt,
	); err != nil {
		return nil, err
	}

	return s, nil
}
