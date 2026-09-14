package config

import (
	"github.com/Nurlan270/cloud-storage-go/internal/auth_server/session"
	"github.com/Nurlan270/cloud-storage-go/internal/core/database"
	"github.com/Nurlan270/cloud-storage-go/internal/core/redis"
)

type Config struct {
	App     App             `mapstructure:"app"      validate:"required"`
	DB      database.Config `mapstructure:"database" validate:"required"`
	Redis   redis.Config    `mapstructure:"redis"    validate:"required"`
	Session session.Config  `mapstructure:"session"  validate:"required"`
}

type App struct {
	Env  string `env:"APP_ENV,required"  mapstructure:"env"  validate:"required"`
	Name string `env:"APP_NAME,required" mapstructure:"name" validate:"required"`
}

func (c Config) GetAppEnv() string {
	return c.App.Env
}

func (c Config) GetAppName() string {
	return c.App.Name
}
