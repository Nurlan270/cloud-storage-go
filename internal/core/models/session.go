package models

import "time"

type Session struct {
	UUID      string `redis:"uuid"`
	UserID    uint64 `redis:"user_id"`
	ExpiresIn time.Duration
}
