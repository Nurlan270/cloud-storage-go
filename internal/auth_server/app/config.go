package app

import (
	"github.com/Nurlan270/cloud-storage-go/internal/core/database"
)

type Config struct {
	Env  string          `env:"APP_ENV,required"  mapstructure:"env"      validate:"required"`
	Name string          `env:"APP_NAME,required" mapstructure:"name"     validate:"required"`
	DB   database.Config `env:",prefix=DB_"       mapstructure:"database" validate:"required"`
}

func (c Config) GetEnv() string {
	return c.Env
}
