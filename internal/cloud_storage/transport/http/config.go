package http

import "time"

type Config struct {
	Address     string        `mapstructure:"address"      validate:"required"`
	Timeout     time.Duration `mapstructure:"timeout"      validate:"required"`
	IdleTimeout time.Duration `mapstructure:"idle_timeout" validate:"required"`
}
