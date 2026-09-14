package redis

import "fmt"

type Config struct {
	Host     string `mapstructure:"host"     validate:"required"`
	Port     uint16 `mapstructure:"port"     validate:"required"`
	Password string `mapstructure:"password" validate:"required" env:"REDIS_PASSWORD,required"`
}

func (c *Config) GetAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}
