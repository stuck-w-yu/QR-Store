package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseVersion(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		want    int64
		wantErr bool
	}{
		{"standard", "000001_init_schema.up.sql", 1, false},
		{"three digits", "000003_phase1_cleanup.up.sql", 3, false},
		{"no underscore", "000001.up.sql", 0, true},
		{"non numeric", "abc_init.up.sql", 0, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseVersion(tc.input)
			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestLoadMigrationsSortedAscending(t *testing.T) {
	migs, err := loadMigrations()
	require.NoError(t, err)
	require.NotEmpty(t, migs, "at least one embedded migration is expected")

	for i, m := range migs {
		assert.NotEmpty(t, m.sql, "migration %s must have SQL content", m.name)
		if i > 0 {
			assert.Greater(t, m.version, migs[i-1].version,
				"migrations must be strictly ascending by version")
		}
	}

	// Version 1 must be the base schema migration.
	assert.Equal(t, int64(1), migs[0].version)

	// The base schema must define the core tables, including those that
	// previously lived only in schema.sql (drift guard).
	base := migs[0].sql
	for _, table := range []string{
		"restaurants", "users", "tables", "orders", "order_items", "payments",
		"service_requests", "payment_webhook_events",
		"registers", "cashier_shifts", "cashier_shift_transactions",
		"shift_closings", "refunds",
	} {
		assert.Containsf(t, base, "CREATE TABLE IF NOT EXISTS "+table, "base schema must create table %s", table)
	}
	// proof_url columns must exist on orders and payments.
	assert.Contains(t, base, "proof_url", "base schema must include proof_url columns")
}
