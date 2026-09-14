package http

import "time"

type Config struct {
	Address      string        `mapstructure:"address"       validate:"required"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"  validate:"required"`
	WriteTimeout time.Duration `mapstructure:"write_timeout" validate:"required"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout"  validate:"required"`
}
