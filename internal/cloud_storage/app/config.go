package app

import (
	"time"

	"github.com/Nurlan270/cloud-storage-go/internal/cloud_storage/transport/http"
	"github.com/Nurlan270/cloud-storage-go/internal/core/database"
)

type Config struct {
	Env        string          `env:"APP_ENV,required"  mapstructure:"env"         validate:"required"`
	Name       string          `env:"APP_NAME,required" mapstructure:"name"        validate:"required"`
	DB         database.Config `env:",prefix=DB_"       mapstructure:"database"    validate:"required"`
	HTTPServer http.Config     `                        mapstructure:"http_server" validate:"required"`
	Session    Session         `                        mapstructure:"session"     validate:"required"`
}

type Session struct {
	ExpiresIn time.Duration `mapstructure:"expires_in" validate:"required"`
}

func (c Config) GetEnv() string {
	return c.Env
}
