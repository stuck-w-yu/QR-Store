//go:build e2e

// Package e2e contains integration tests that require a real PostgreSQL
// database. They are compiled only with the `e2e` build tag and skipped unless
// TEST_DATABASE_URL is set, so the default `go test ./...` stays hermetic.
//
// Locally:
//   TEST_DATABASE_URL=postgres://postgres:postgres@localhost:5432/qr_test?sslmode=disable \
//   go test -tags=e2e ./tests/e2e/...
//
// In CI a Postgres service container provides the database (see .github/workflows/ci.yml).
package e2e

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
	"qr-store/backend/pkg/database"
	"qr-store/backend/pkg/logger"
)

func TestMain(m *testing.M) {
	logger.Init("test")
	os.Exit(m.Run())
}

// testDB connects to TEST_DATABASE_URL, wipes the schema, runs migrations, and
// returns the pool. It skips the test if no database is configured.
func testDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set; skipping e2e test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	require.NoError(t, err, "connect to test database")
	require.NoError(t, pool.Ping(ctx), "ping test database")

	// Start from a clean slate for deterministic runs.
	_, err = pool.Exec(ctx, `DROP SCHEMA public CASCADE; CREATE SCHEMA public;`)
	require.NoError(t, err, "reset schema")

	db := &database.DB{Pool: pool}
	require.NoError(t, database.RunMigrations(ctx, pool), "run migrations")

	t.Cleanup(pool.Close)

	// Sanity: the runner must be idempotent.
	require.NoError(t, database.RunMigrations(ctx, pool), "second RunMigrations must be a no-op")

	_ = db
	return pool
}
