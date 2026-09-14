package minio

type Config struct {
	User     string `env:"MINIO_USER,required"     mapstructure:"user"     validate:"required"`
	Password string `env:"MINIO_PASSWORD,required" mapstructure:"password" validate:"required"`
}
