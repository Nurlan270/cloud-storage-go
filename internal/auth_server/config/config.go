package config

import (
	"github.com/Nurlan270/cloud-storage-go/internal/auth_server/session"
	"github.com/Nurlan270/cloud-storage-go/internal/core/database"
	"github.com/Nurlan270/cloud-storage-go/internal/core/redis"
)

type Config struct {
	App     App             `env:",prefix=APP_"   mapstructure:"app"      validate:"required"`
	DB      database.Config `env:",prefix=DB_"    mapstructure:"database" validate:"required"`
	Redis   redis.Config    `env:",prefix=REDIS_" mapstructure:"redis"    validate:"required"`
	Session session.Config  `                     mapstructure:"session"  validate:"required"`
}

type App struct {
	Env  string `env:"ENV,required"  mapstructure:"env"  validate:"required"`
	Name string `env:"NAME,required" mapstructure:"name" validate:"required"`
}

func (c Config) GetAppEnv() string {
	return c.App.Env
}

func (c Config) GetAppName() string {
	return c.App.Name
}
