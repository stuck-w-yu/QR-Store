//go:build e2e

package e2e

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"qr-store/backend/internal/cashier"
	"qr-store/backend/internal/order"
)

// seedRestaurant inserts a minimal restaurant row.
func seedRestaurant(t *testing.T, ctx context.Context, pool *pgxpool.Pool, id, slug string) {
	t.Helper()
	now := time.Now()
	_, err := pool.Exec(ctx, `
		INSERT INTO restaurants (id, name, slug, tax_percent, service_percent, status, created_at, updated_at)
		VALUES ($1, $2, $3, 10.00, 0.00, 'ACTIVE', $4, $4)
	`, id, "Resto "+id, slug, now)
	require.NoError(t, err)
}

func seedRegister(t *testing.T, ctx context.Context, pool *pgxpool.Pool, id, restaurantID, name string) {
	t.Helper()
	now := time.Now()
	_, err := pool.Exec(ctx, `
		INSERT INTO registers (id, restaurant_id, outlet_id, name, status, created_at, updated_at)
		VALUES ($1, $2, NULL, $3, 'ACTIVE', $4, $4)
	`, id, restaurantID, name, now)
	require.NoError(t, err)
}

func TestRegister_TenantIsolation(t *testing.T) {
	pool := testDB(t)
	ctx := context.Background()

	seedRestaurant(t, ctx, pool, "rst_a", "resto-a")
	seedRestaurant(t, ctx, pool, "rst_b", "resto-b")
	seedRegister(t, ctx, pool, "reg_a", "rst_a", "Kasir A")

	repo := cashier.NewRepository(pool)

	// Owning tenant can fetch it.
	reg, err := repo.GetRegisterByID(ctx, "reg_a", "rst_a")
	require.NoError(t, err)
	assert.Equal(t, "reg_a", reg.ID)
	assert.Equal(t, "rst_a", reg.RestaurantID)

	// A different tenant must NOT be able to fetch it (IDOR prevention).
	_, err = repo.GetRegisterByID(ctx, "reg_a", "rst_b")
	require.ErrorIs(t, err, cashier.ErrRegisterNotFound,
		"cross-tenant register access must return not-found, not the record")

	// A different tenant must NOT be able to delete it either.
	err = repo.DeleteRegister(ctx, "reg_a", "rst_b")
	require.ErrorIs(t, err, cashier.ErrRegisterNotFound)

	// The register must still exist and be deletable by its owner.
	reg, err = repo.GetRegisterByID(ctx, "reg_a", "rst_a")
	require.NoError(t, err)
	assert.NotNil(t, reg)
}

func TestOrder_TenantScopedDelete(t *testing.T) {
	pool := testDB(t)
	ctx := context.Background()

	seedRestaurant(t, ctx, pool, "rst_a", "resto-a")
	seedRestaurant(t, ctx, pool, "rst_b", "resto-b")

	// Seed tables (orders require a valid table_id FK).
	now := time.Now()
	for _, tbl := range []struct{ id, resto string }{{"tbl_a", "rst_a"}, {"tbl_b", "rst_b"}} {
		_, err := pool.Exec(ctx, `
			INSERT INTO tables (id, restaurant_id, name, qr_token, status, created_at, updated_at)
			VALUES ($1, $2, $3, $4, 'ACTIVE', $5, $5)
		`, tbl.id, tbl.resto, "Meja "+tbl.id, "qr_"+tbl.id, now)
		require.NoError(t, err)
	}

	repo := order.NewRepository(pool)
	o := &order.Order{
		ID:           "ord_a1",
		RestaurantID: "rst_a",
		TableID:      "tbl_a",
		OrderNumber:  "ORD-E2E-0001",
		Status:       order.StatusWaitingPayment,
		Subtotal:     10000,
		Tax:          1000,
		Total:        11000,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	require.NoError(t, repo.Create(ctx, o, nil))

	// Wrong tenant cannot delete.
	err := repo.Delete(ctx, "ord_a1", "rst_b")
	require.ErrorIs(t, err, order.ErrOrderNotFound)

	// Correct tenant succeeds.
	require.NoError(t, repo.Delete(ctx, "ord_a1", "rst_a"))

	_, err = repo.GetByID(ctx, "ord_a1")
	require.ErrorIs(t, err, order.ErrOrderNotFound)
}

func TestOrder_NumberCollisionReturnsConflict(t *testing.T) {
	pool := testDB(t)
	ctx := context.Background()

	seedRestaurant(t, ctx, pool, "rst_a", "resto-a")
	now := time.Now()
	_, err := pool.Exec(ctx, `
		INSERT INTO tables (id, restaurant_id, name, qr_token, status, created_at, updated_at)
		VALUES ('tbl_a', 'rst_a', 'Meja A', 'qr_a', 'ACTIVE', $1, $1)
	`, now)
	require.NoError(t, err)

	repo := order.NewRepository(pool)

	mk := func(id string) *order.Order {
		return &order.Order{
			ID:           id,
			RestaurantID: "rst_a",
			TableID:      "tbl_a",
			OrderNumber:  "ORD-DUPLICATE",
			Status:       order.StatusWaitingPayment,
			Subtotal:     5000,
			Tax:          500,
			Total:        5500,
			CreatedAt:    now,
			UpdatedAt:    now,
		}
	}

	require.NoError(t, repo.Create(ctx, mk("ord_first"), nil))

	// Second insert with the same order_number must surface the sentinel error so
	// the service can retry with a fresh number.
	err = repo.Create(ctx, mk("ord_second"), nil)
	require.ErrorIs(t, err, order.ErrOrderNumberConflict)
}

func TestOrder_ListItemErrorNotSwallowed(t *testing.T) {
	pool := testDB(t)
	ctx := context.Background()

	seedRestaurant(t, ctx, pool, "rst_a", "resto-a")
	now := time.Now()
	_, err := pool.Exec(ctx, `
		INSERT INTO tables (id, restaurant_id, name, qr_token, status, created_at, updated_at)
		VALUES ('tbl_a', 'rst_a', 'Meja A', 'qr_a', 'ACTIVE', $1, $1)
	`, now)
	require.NoError(t, err)

	repo := order.NewRepository(pool)
	o := &order.Order{
		ID: "ord_items", RestaurantID: "rst_a", TableID: "tbl_a",
		OrderNumber: "ORD-ITEMS", Status: order.StatusWaitingPayment,
		Subtotal: 1000, Tax: 100, Total: 1100, CreatedAt: now, UpdatedAt: now,
	}
	require.NoError(t, repo.Create(ctx, o, []order.OrderItem{{
		ID: "oi_1", OrderID: "ord_items", MenuNameSnapshot: "Nasi Goreng",
		UnitPrice: 1000, Quantity: 1, Subtotal: 1000, CreatedAt: now,
	}}))

	got, err := repo.GetByID(ctx, "ord_items")
	require.NoError(t, err)
	require.Len(t, got.Items, 1, "order items must be loaded (error not swallowed)")
	assert.Equal(t, "Nasi Goreng", got.Items[0].MenuNameSnapshot)
}
