package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

func MustConnect(conf Config, db int) *redis.Client {
	rdb, err := connect(conf, db)
	if err != nil {
		panic(err)
	}

	return rdb
}

func Connect(conf Config, db int) (*redis.Client, error) {
	return connect(conf, db)
}

func connect(conf Config, db int) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     conf.GetAddr(),
		Password: conf.Password,
		DB:       db,
	})

	_, err := rdb.Ping(context.Background()).Result()
	if err != nil {
		return nil, fmt.Errorf("redis: failed to connect: %s", err)
	}

	return rdb, nil
}
