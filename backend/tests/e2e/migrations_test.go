//go:build e2e

package e2e

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMigrations_CreateCoreTables(t *testing.T) {
	pool := testDB(t)
	ctx := context.Background()

	for _, table := range []string{
		"restaurants", "users", "tables", "table_sessions", "categories", "menus",
		"menu_modifiers", "menu_modifier_options", "orders", "order_items", "payments",
		"service_requests", "payment_webhook_events", "order_status_history", "audit_logs",
		"registers", "cashier_shifts", "cashier_shift_transactions", "shift_closings", "refunds",
		"schema_migrations",
	} {
		var exists bool
		err := pool.QueryRow(ctx, `
			SELECT EXISTS (
				SELECT 1 FROM information_schema.tables
				WHERE table_schema = 'public' AND table_name = $1
			)
		`, table).Scan(&exists)
		require.NoError(t, err)
		assert.Truef(t, exists, "expected table %q to exist after migrations", table)
	}
}

func TestMigrations_RecordsVersions(t *testing.T) {
	pool := testDB(t)
	ctx := context.Background()

	var count int
	require.NoError(t, pool.QueryRow(ctx, `SELECT COUNT(*) FROM schema_migrations`).Scan(&count))
	assert.GreaterOrEqual(t, count, 1, "at least one migration must be recorded")

	// The base schema (version 1) must be present.
	var baseExists bool
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = 1)`).Scan(&baseExists))
	assert.True(t, baseExists)
}

func TestMigrations_OrderNumberHasUniqueConstraint(t *testing.T) {
	pool := testDB(t)
	ctx := context.Background()

	// order_number must be globally unique so the retry-on-conflict logic works.
	var isUnique bool
	err := pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM pg_indexes
			WHERE schemaname = 'public' AND tablename = 'orders'
			  AND indexdef ILIKE '%UNIQUE%order_number%'
		)
	`).Scan(&isUnique)
	require.NoError(t, err)
	assert.True(t, isUnique, "orders.order_number must have a UNIQUE constraint/index")
}
