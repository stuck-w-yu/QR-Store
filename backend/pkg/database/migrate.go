package database

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"qr-store/backend/pkg/logger"
)

// migrationsFS embeds the ordered, numbered SQL migrations. These files are the
// single source of truth for the database schema; schema.sql is no longer used.
//
//go:embed migrations/*.up.sql
var migrationsFS embed.FS

// migration is a single numbered .up.sql file.
type migration struct {
	version int64
	name    string
	sql     string
}

// loadMigrations reads and sorts all embedded *.up.sql files by version.
func loadMigrations() ([]migration, error) {
	entries, err := fs.ReadDir(migrationsFS, "migrations")
	if err != nil {
		return nil, fmt.Errorf("read migrations dir: %w", err)
	}

	var migs []migration
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".up.sql") {
			continue
		}
		version, err := parseVersion(e.Name())
		if err != nil {
			return nil, err
		}
		content, err := migrationsFS.ReadFile(path.Join("migrations", e.Name()))
		if err != nil {
			return nil, fmt.Errorf("read migration %s: %w", e.Name(), err)
		}
		migs = append(migs, migration{version: version, name: e.Name(), sql: string(content)})
	}

	sort.Slice(migs, func(i, j int) bool { return migs[i].version < migs[j].version })
	return migs, nil
}

// parseVersion extracts the leading numeric prefix from a migration filename
// such as "000001_init_schema.up.sql".
func parseVersion(name string) (int64, error) {
	idx := strings.IndexByte(name, '_')
	if idx <= 0 {
		return 0, fmt.Errorf("migration %q does not start with a numeric version", name)
	}
	v, err := strconv.ParseInt(name[:idx], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("migration %q has an invalid version prefix: %w", name, err)
	}
	return v, nil
}

// ensureMigrationsTable creates the bookkeeping table used to track applied
// migrations.
func ensureMigrationsTable(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version     BIGINT PRIMARY KEY,
			name        VARCHAR(255) NOT NULL,
			applied_at  TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL
		)
	`)
	return err
}

// appliedVersions returns the set of already-applied migration versions.
func appliedVersions(ctx context.Context, pool *pgxpool.Pool) (map[int64]bool, error) {
	rows, err := pool.Query(ctx, `SELECT version FROM schema_migrations`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[int64]bool)
	for rows.Next() {
		var v int64
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		applied[v] = true
	}
	return applied, rows.Err()
}

// isLegacyDatabase reports whether the database already contains the core
// schema (e.g. created by an older version that ran schema.sql directly) but
// has no migration bookkeeping. Such databases are baselined: existing
// migrations are marked as applied without re-running them.
func isLegacyDatabase(ctx context.Context, pool *pgxpool.Pool) (bool, error) {
	var exists bool
	err := pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.tables
			WHERE table_schema = 'public' AND table_name = 'restaurants'
		)
	`).Scan(&exists)
	return exists, err
}

// RunMigrations applies all pending numbered migrations exactly once, tracked
// in the schema_migrations table. It is safe to run on every boot.
func RunMigrations(ctx context.Context, pool *pgxpool.Pool) error {
	migs, err := loadMigrations()
	if err != nil {
		return err
	}
	if len(migs) == 0 {
		return fmt.Errorf("no migrations found")
	}

	if err := ensureMigrationsTable(ctx, pool); err != nil {
		return fmt.Errorf("ensure schema_migrations: %w", err)
	}

	applied, err := appliedVersions(ctx, pool)
	if err != nil {
		return fmt.Errorf("read applied migrations: %w", err)
	}

	// Baseline a pre-existing (legacy) database so we don't replay destructive
	// or redundant DDL. All currently known migrations are marked applied.
	if len(applied) == 0 {
		legacy, err := isLegacyDatabase(ctx, pool)
		if err != nil {
			return fmt.Errorf("detect legacy database: %w", err)
		}
		if legacy {
			logger.Log.Warn("legacy database detected without migration history; baselining existing schema")
			for _, m := range migs {
				if _, err := pool.Exec(ctx,
					`INSERT INTO schema_migrations (version, name) VALUES ($1, $2) ON CONFLICT (version) DO NOTHING`,
					m.version, m.name,
				); err != nil {
					return fmt.Errorf("baseline migration %d: %w", m.version, err)
				}
			}
			return nil
		}
	}

	for _, m := range migs {
		if applied[m.version] {
			continue
		}
		logger.Log.Info("applying migration", "version", m.version, "name", m.name)

		// Each migration runs in its own transaction so a failure leaves the
		// database in the last successfully-migrated state.
		tx, err := pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("begin tx for migration %d: %w", m.version, err)
		}
		if _, err := tx.Exec(ctx, m.sql, pgx.QueryExecModeSimpleProtocol); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("migration %d (%s) failed: %w", m.version, m.name, err)
		}
		if _, err := tx.Exec(ctx,
			`INSERT INTO schema_migrations (version, name, applied_at) VALUES ($1, $2, $3)`,
			m.version, m.name, time.Now(),
		); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("record migration %d: %w", m.version, err)
		}
		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("commit migration %d: %w", m.version, err)
		}
	}

	logger.Log.Info("database migrations up to date", "latest_version", migs[len(migs)-1].version)
	return nil
}
