package testutil

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/Nurlan270/cloud-storage-go/internal/auth_server/testutil/closer"
	"github.com/jackc/pgx/v5/pgxpool"
	"net"
	"path/filepath"
	"testing"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"

	"github.com/pressly/goose/v3"

	"github.com/Nurlan270/cloud-storage-go/internal/core/database"
)

var migrationsDir = filepath.Join("..", "..", "..", "core", "database", "migrations")

func NewTestDB(ctx context.Context, conf *database.Config) (*sql.DB, error) {
	pgc, err := postgres.Run(ctx,
		"postgres:18-alpine",
		postgres.WithDatabase(conf.Name),
		postgres.WithUsername(conf.Username),
		postgres.WithPassword(conf.Password),
		postgres.BasicWaitStrategies(),
	)
	if err != nil {
		return nil, fmt.Errorf("postgres: failed to start container: %s", err)
	}

	//	Get Host/Port pair from container
	dbHost, dbPort, err := splitEndpoint(ctx, pgc)
	if err != nil {
		return nil, fmt.Errorf("failed to parse postgres host and port: %s", err)
	}

	conf.Host = dbHost
	conf.Port = dbPort

	db, err := database.ConnectStd(*conf)
	if err != nil {
		return nil, fmt.Errorf("db: failed to connect: %s", err)
	}

	closer.Add("Test Database", func() error {
		dbErr := db.Close()
		return errors.Join(dbErr, terminateTestDB(pgc))
	})

	//	Setup Goose
	if err = goose.SetDialect("postgres"); err != nil {
		return nil, fmt.Errorf("goose: failed to set dialect: %s", err)
	}

	//	Run migrations
	if err = goose.Up(db, migrationsDir); err != nil {
		return nil, fmt.Errorf("goose: failed to run migrations: %s", err)
	}

	return db, nil
}

func NewTestDBPool(conf database.Config) (*pgxpool.Pool, error) {
	pool, err := database.Connect(conf)
	if err != nil {
		return nil, fmt.Errorf("pool: failed to connect: %s", err)
	}

	closer.Add("Test Pool", func() error {
		pool.Close()
		return nil
	})

	return pool, nil
}

func CleanupDatabase(t *testing.T, db *sql.DB) {
	t.Helper()

	err := goose.Reset(db, migrationsDir)
	if err != nil {
		t.Fatalf("db: failed to reset migrations: %s", err)
	}

	err = goose.Up(db, migrationsDir)
	if err != nil {
		t.Fatalf("goose: failed to run migrations: %s", err)
	}
}

func splitEndpoint(ctx context.Context, pgc *postgres.PostgresContainer) (host, port string, err error) {
	hostPort, err := pgc.Endpoint(ctx, "")
	if err != nil {
		return "", "", fmt.Errorf("failed to retrieve postgres host/port pair: %s", err)
	}

	host, port, err = net.SplitHostPort(hostPort)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse postgres host and port: %s", err)
	}

	return host, port, nil
}

func terminateTestDB(pgc *postgres.PostgresContainer) error {
	if err := testcontainers.TerminateContainer(pgc); err != nil {
		return fmt.Errorf("postgres: failed to terminate container: %s", err)
	}

	return nil
}
