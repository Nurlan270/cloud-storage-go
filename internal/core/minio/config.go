package minio

import (
	"fmt"
)

type Config struct {
	Host     string `mapstructure:"host"     validate:"required"`
	Port     uint16 `mapstructure:"port"     validate:"required"`
	User     string `mapstructure:"user"     validate:"required" env:"MINIO_USER,required"`
	Password string `mapstructure:"password" validate:"required" env:"MINIO_PASSWORD,required"`
}

func (c Config) GetAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}
