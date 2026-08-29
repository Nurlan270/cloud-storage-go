package database

type Config struct {
	Host     string `mapstructure:"host"     validate:"required"`
	Port     string `mapstructure:"port"     validate:"required"`
	Name     string `mapstructure:"name"     validate:"required"`
	Username string `mapstructure:"username" validate:"required" env:"USERNAME,required"`
	Password string `mapstructure:"password" validate:"required" env:"PASSWORD,required"`
}
