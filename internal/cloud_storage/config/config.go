package config

import (
	"github.com/Nurlan270/cloud-storage-go/internal/cloud_storage/transport/http"
	"github.com/Nurlan270/cloud-storage-go/internal/core/database"
)

type Config struct {
	App        App             `env:",prefix=APP_" mapstructure:"app"         validate:"required"`
	DB         database.Config `env:",prefix=DB_"  mapstructure:"database"    validate:"required"`
	HTTPServer http.Config     `                   mapstructure:"http_server" validate:"required"`
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
