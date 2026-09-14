package repository

import (
	"context"

	"github.com/redis/go-redis/v9"

	errs "github.com/Nurlan270/cloud-storage-go/internal/core/errors"
	"github.com/Nurlan270/cloud-storage-go/internal/core/models"
)

type SessionRepository interface {
	CreateSession(session *models.Session) (*models.Session, error)
	GetSessionFromSID(sid string) (*models.Session, error)
	DeleteSession(sid string) error
}

type sessionRepository struct {
	rdb *redis.Client
}

func NewSessionRepository(rdb *redis.Client) SessionRepository {
	return sessionRepository{rdb: rdb}
}

func (r sessionRepository) CreateSession(session *models.Session) (*models.Session, error) {
	var key = "sessions:" + session.UUID

	ctx := context.Background()
	pipe := r.rdb.Pipeline()

	//	Execute both in pipe for atomicity
	pipe.HSet(ctx, key, session)
	pipe.Expire(ctx, key, session.ExpiresIn)

	if _, err := pipe.Exec(ctx); err != nil {
		return nil, err
	}

	return session, nil
}

func (r sessionRepository) GetSessionFromSID(sid string) (*models.Session, error) {
	var key = "sessions:" + sid

	var session models.Session
	if err := r.rdb.HGetAll(context.Background(), key).Scan(&session); err != nil {
		return nil, err
	}

	if session.UUID == "" {
		//	Session not found or was expired by TTL
		return nil, errs.ErrSessionInvalid
	}

	return &session, nil
}

func (r sessionRepository) DeleteSession(sid string) error {
	var key = "sessions:" + sid

	if err := r.rdb.Del(context.Background(), key).Err(); err != nil {
		return err
	}

	return nil
}
