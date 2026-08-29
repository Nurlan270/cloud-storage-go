package config

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/sethvargo/go-envconfig"
	"github.com/spf13/viper"

	gpv "github.com/go-playground/validator/v10"
)

type Config interface {
	GetEnv() string
}

func MustLoad[C Config]() *C {
	if err := godotenv.Load(); err != nil {
		panic(fmt.Sprintf("failed to load .env file: %s", err))
	}

	var conf C

	ctx := context.Background()

	if err := envconfig.Process(ctx, &conf); err != nil {
		panic(fmt.Sprintf("failed to proccess .env file: %s", err))
	}

	//	Viper setup
	v := viper.New()
	v.SetConfigFile(filepath.Join("config", conf.GetEnv()+".yml"))

	//	Read config file
	if err := v.ReadInConfig(); err != nil {
		panic(fmt.Sprintf("viper: failed reading config: %s", err))
	}

	//	Unmarshal config file's data into Config struct
	if err := v.Unmarshal(&conf); err != nil {
		panic(fmt.Sprintf("viper: failed to decode into struct: %s", err))
	}

	validate := gpv.New(gpv.WithRequiredStructEnabled())

	//	Validate final config
	if err := validate.Struct(conf); err != nil {
		panic(fmt.Sprintf("validator: failed to validate config: %s", err))
	}

	return &conf
}
