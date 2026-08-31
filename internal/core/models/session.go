package models

import "time"

type Session struct {
	UUID      string
	UserID    uint64
	ExpiresAt time.Time
}
