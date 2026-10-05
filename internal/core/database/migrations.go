package database

import (
	"database/sql"
	"embed"
	"fmt"

	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func RunMigrations(db *sql.DB) error {
	goose.SetBaseFS(migrationsFS)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("goose: failed to set dialect: %s", err)
	}

	if err := goose.Up(db, "migrations"); err != nil {
		return fmt.Errorf("goose: failed to run migrations: %w", err)
	}

	return nil
}

func ResetMigrations(db *sql.DB) error {
	goose.SetBaseFS(migrationsFS)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("goose: failed to set dialect: %s", err)
	}

	if err := goose.Reset(db, "migrations"); err != nil {
		return fmt.Errorf("goose: failed to reset migrations: %w", err)
	}

	return nil
}
