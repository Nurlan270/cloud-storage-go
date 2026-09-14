package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func MustConnect(conf Config) *pgxpool.Pool {
	pool, err := connect(conf)
	if err != nil {
		panic(err)
	}

	return pool
}

func Connect(conf Config) (*pgxpool.Pool, error) {
	return connect(conf)
}

func connect(conf Config) (*pgxpool.Pool, error) {
	connString := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable pool_max_conns=25",
		conf.Host, conf.Port, conf.Username, conf.Password, conf.Name,
	)

	pool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		return nil, fmt.Errorf("postgres: failed to create pool: %s", err)
	}

	return pool, nil
}
