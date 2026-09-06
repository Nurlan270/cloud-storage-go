package testutil

import (
	"time"

	"github.com/Nurlan270/cloud-storage-go/internal/auth_server/config"
	"github.com/Nurlan270/cloud-storage-go/internal/auth_server/session"
	"github.com/Nurlan270/cloud-storage-go/internal/core/database"
	"github.com/Nurlan270/cloud-storage-go/internal/core/redis"
)

func NewTestConfig() *config.Config {
	app := config.App{
		Env:  "testing",
		Name: "testing",
	}

	db := database.Config{
		Name:     "testing",
		Username: "postgres",
		Password: "postgres",
	}

	rdb := redis.Config{}

	sess := session.Config{
		ExpiresIn: 5 * time.Minute,
	}

	return &config.Config{
		App:     app,
		DB:      db,
		Redis:   rdb,
		Session: sess,
	}
}
