package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"qr-store/backend/pkg/logger"
)

type DB struct {
	Pool *pgxpool.Pool
}

func NewPostgresPool(ctx context.Context, connString string) (*DB, error) {
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("unable to parse database config: %w", err)
	}

	config.MaxConns = 25
	config.MinConns = 5
	config.MaxConnLifetime = 1 * time.Hour
	config.MaxConnIdleTime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	// Retry with exponential backoff: the database may still be starting up,
	// especially under Docker Compose / Kubernetes where the app can boot
	// before PostgreSQL accepts connections.
	const (
		maxAttempts = 10
		baseDelay   = 500 * time.Millisecond
		maxDelay    = 10 * time.Second
	)
	delay := baseDelay
	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		lastErr = pool.Ping(pingCtx)
		cancel()
		if lastErr == nil {
			return &DB{Pool: pool}, nil
		}

		logger.Log.Warn("database ping failed, retrying",
			"attempt", attempt, "max_attempts", maxAttempts, "delay", delay, "error", lastErr)

		if attempt == maxAttempts {
			break
		}
		select {
		case <-ctx.Done():
			pool.Close()
			return nil, fmt.Errorf("context cancelled while waiting for database: %w", ctx.Err())
		case <-time.After(delay):
		}
		delay *= 2
		if delay > maxDelay {
			delay = maxDelay
		}
	}

	pool.Close()
	return nil, fmt.Errorf("unable to ping database after %d attempts: %w", maxAttempts, lastErr)
}

func (db *DB) Close() {
	if db.Pool != nil {
		db.Pool.Close()
	}
}
