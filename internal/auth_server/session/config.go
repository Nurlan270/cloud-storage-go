package session

import "time"

type Config struct {
	ExpiresIn time.Duration `mapstructure:"expires_in" validate:"required"`
}
