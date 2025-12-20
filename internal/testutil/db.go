package testutil

import (
	"context"
	utildb "github.com/gobugger/gomarket/internal/util/db"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"path/filepath"
)

// Starts postgres container, runs migrations and returns connected db handle
func SetupDB() (db *pgxpool.Pool, cleanup func(), dsn string) {
	// Create test-container to run postgres
	ctx := context.Background()
	postgresContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("user"),
		postgres.WithPassword("password"),
		postgres.BasicWaitStrategies(),
	)
	if err != nil {
		panic(err)
	}

	dsn, err = postgresContainer.ConnectionString(ctx)
	if err != nil {
		panic(err)
	}

	dsn += "sslmode=disable"

	root, err := FindProjectRoot()
	if err != nil {
		panic(err)
	}

	// Migrate
	m, err := migrate.New("file://"+filepath.Join(root, "migrations"), dsn)
	if err != nil {
		panic(err)
	}

	if err := m.Up(); err != nil {
		panic(err)
	}

	pool, err := utildb.Open(ctx, dsn)
	if err != nil {
		panic(err)
	}

	cleanup = func() {
		pool.Close()
		if err := testcontainers.TerminateContainer(postgresContainer); err != nil {
			panic(err)
		}
	}

	return pool, cleanup, dsn
}

// Migrate down and back up
func CleanDB(dsn string) error {
	root, err := FindProjectRoot()
	if err != nil {
		panic(err)
	}

	// Migrate
	m, err := migrate.New("file://"+filepath.Join(root, "migrations"), dsn)
	if err != nil {
		panic(err)
	}

	if err := m.Down(); err != nil {
		panic(err)
	}

	if err := m.Up(); err != nil {
		panic(err)
	}

	return nil
}
