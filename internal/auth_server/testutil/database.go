package testutil

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"path/filepath"
	"testing"

	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"

	"github.com/Nurlan270/cloud-storage-go/internal/core/database"
)

func NewTestDB(ctx context.Context, conf database.Config) (*sql.DB, func() error, error) {
	pgc, err := postgres.Run(ctx,
		"postgres:18-alpine",
		postgres.WithDatabase(conf.Name),
		postgres.WithUsername(conf.Username),
		postgres.WithPassword(conf.Password),
		postgres.BasicWaitStrategies(),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("postgres: failed to start container: %s", err)
	}

	//	Get Host/Port pair from container
	dbHostPort, err := pgc.Endpoint(ctx, "")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to retrieve postgres host/port pair: %s", err)
	}

	dbHost, dbPort, err := net.SplitHostPort(dbHostPort)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to parse postgres host and port: %s", err)
	}

	conf.Host = dbHost
	conf.Port = dbPort

	db, err := database.Connect(conf)
	if err != nil {
		return nil, nil, fmt.Errorf("db: failed to connect: %s", err)
	}

	cleanup := func() error {
		return errors.Join(
			db.Close(),
			terminateTestDB(pgc),
		)
	}

	//	Setup Goose
	if err = goose.SetDialect("postgres"); err != nil {
		return nil, cleanup, fmt.Errorf("goose: failed to set dialect: %s", err)
	}

	//	Run migrations
	if err = goose.Up(db, filepath.Join("..", "..", "..", "core", "database", "migrations")); err != nil {
		return nil, cleanup, fmt.Errorf("goose: failed to run migrations: %s", err)
	}

	return db, cleanup, nil
}

func CleanupDatabase(t *testing.T, testDB *sql.DB) {
	t.Helper()

	_, err := testDB.Exec(`
		TRUNCATE TABLE sessions, users RESTART IDENTITY CASCADE
	`)
	if err != nil {
		t.Fatalf("db: failed to truncate tables: %s", err)
	}
}

func terminateTestDB(pgc *postgres.PostgresContainer) error {
	if err := testcontainers.TerminateContainer(pgc); err != nil {
		return fmt.Errorf("postgres: failed to terminate container: %s", err)
	}

	return nil
}
