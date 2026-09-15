package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

func MustConnect(conf Config) *pgxpool.Pool {
	pool, err := connect(conf)
	if err != nil {
		panic(err)
	}

	return pool
}

func Connect(conf Config) (*pgxpool.Pool, error) {
	pool, err := connect(conf)
	if err != nil {
		return nil, fmt.Errorf("postgres: failed to connect: %s", err)
	}

	return pool, nil
}

func ConnectStd(conf Config) (*sql.DB, error) {
	pool, err := connect(conf)
	if err != nil {
		return nil, fmt.Errorf("postgres: failed to connect: %s", err)
	}

	return stdlib.OpenDBFromPool(pool), nil
}

func connect(conf Config) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(context.Background(), conf.ConnString())
	if err != nil {
		return nil, fmt.Errorf("postgres: failed to create pool: %s", err)
	}

	return pool, nil
}
