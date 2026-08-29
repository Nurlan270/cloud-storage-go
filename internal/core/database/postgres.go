package database

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

func MustConnect(conf Config) *sql.DB {
	connString := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		conf.Host, conf.Port, conf.Username, conf.Password, conf.Name,
	)

	db, err := sql.Open("postgres", connString)
	if err != nil {
		panic(fmt.Sprintf("db: failed to open: %s", err))
	}

	if err = db.Ping(); err != nil {
		panic(fmt.Sprintf("db: failed to ping: %s", err))
	}

	return db
}
