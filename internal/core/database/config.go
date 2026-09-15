package database

import "fmt"

type Config struct {
	Host     string `mapstructure:"host"     validate:"required"`
	Port     string `mapstructure:"port"     validate:"required"`
	Name     string `mapstructure:"name"     validate:"required" env:"DB_NAME,required"`
	Username string `mapstructure:"username" validate:"required" env:"DB_USERNAME,required"`
	Password string `mapstructure:"password" validate:"required" env:"DB_PASSWORD,required"`
}

func (c *Config) ConnString() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable pool_max_conns=25",
		c.Host, c.Port, c.Username, c.Password, c.Name,
	)
}
