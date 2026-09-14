package config

import (
	"github.com/Nurlan270/cloud-storage-go/internal/cloud_storage/transport/http"
	"github.com/Nurlan270/cloud-storage-go/internal/core/database"
	"github.com/Nurlan270/cloud-storage-go/internal/core/minio"
	"github.com/Nurlan270/cloud-storage-go/internal/core/redis"
)

type Config struct {
	App        App             `mapstructure:"app"         validate:"required"`
	DB         database.Config `mapstructure:"database"    validate:"required"`
	Redis      redis.Config    `mapstructure:"redis"       validate:"required"`
	Minio      minio.Config    `mapstructure:"minio"       validate:"required"`
	HTTPServer http.Config     `mapstructure:"http_server" validate:"required"`
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
