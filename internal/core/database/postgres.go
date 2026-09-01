package database

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

func MustConnect(conf Config) *sql.DB {
	db, err := connect(conf)
	if err != nil {
		panic(err)
	}

	return db
}

func Connect(conf Config) (*sql.DB, error) {
	return connect(conf)
}

func connect(conf Config) (*sql.DB, error) {
	connString := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		conf.Host, conf.Port, conf.Username, conf.Password, conf.Name,
	)

	db, err := sql.Open("postgres", connString)
	if err != nil {
		return nil, fmt.Errorf("db: failed to open: %s", err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("db: failed to ping: %s", err)
	}

	return db, nil
}
