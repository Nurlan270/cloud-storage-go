package redis

type Config struct {
	Host     string `mapstructure:"host"     validate:"required"`
	Port     uint16 `mapstructure:"port"     validate:"required"`
	Password string `mapstructure:"password" validate:"required" env:"PASSWORD,required"`
}
