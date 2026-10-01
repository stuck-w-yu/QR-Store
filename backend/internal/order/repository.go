package order

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrOrderNotFound      = errors.New("order not found")
	ErrInvalidStatusOrder = errors.New("invalid status transition")
)

type Repository interface {
	Create(ctx context.Context, o *Order, items []OrderItem) error
	GetByID(ctx context.Context, id string) (*Order, error)
	GetByOrderNumber(ctx context.Context, orderNum string) (*Order, error)
	List(ctx context.Context, restaurantID string, status *Status) ([]Order, error)
	ListKitchenOrders(ctx context.Context, restaurantID string) ([]Order, error)
	UpdateStatus(ctx context.Context, id string, newStatus Status, changedBy *string) error
	GetOrderItems(ctx context.Context, orderID string) ([]OrderItem, error)
	GetTodayStats(ctx context.Context, restaurantID string) (revenue int64, orderCount int, avgOrder int64, paidCount int, cancelledCount int, err error)
}

type repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) Repository {
	return &repository{pool: pool}
}

func (r *repository) Create(ctx context.Context, o *Order, items []OrderItem) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	orderQuery := `
		INSERT INTO orders (id, restaurant_id, table_id, table_session_id, order_number, status, subtotal, tax, service_charge, discount, total, notes, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`
	_, err = tx.Exec(ctx, orderQuery,
		o.ID, o.RestaurantID, o.TableID, o.TableSessionID, o.OrderNumber, o.Status,
		o.Subtotal, o.Tax, o.ServiceCharge, o.Discount, o.Total, o.Notes, o.CreatedAt, o.UpdatedAt,
	)
	if err != nil {
		return err
	}

	itemQuery := `
		INSERT INTO order_items (id, order_id, menu_id, menu_name_snapshot, unit_price, quantity, subtotal, selected_modifiers, notes, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	for _, it := range items {
		modsJSON, _ := json.Marshal(it.SelectedModifiers)
		_, err := tx.Exec(ctx, itemQuery,
			it.ID, o.ID, it.MenuID, it.MenuNameSnapshot, it.UnitPrice, it.Quantity, it.Subtotal, modsJSON, it.Notes, it.CreatedAt,
		)
		if err != nil {
			return err
		}
	}

	// Initial status history
	historyQuery := `
		INSERT INTO order_status_history (id, order_id, from_status, to_status, changed_by, created_at)
		VALUES ($1, $2, NULL, $3, $4, $5)
	`
	_, err = tx.Exec(ctx, historyQuery, "osh_"+o.ID, o.ID, o.Status, "CUSTOMER", o.CreatedAt)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *repository) GetByID(ctx context.Context, id string) (*Order, error) {
	query := `
		SELECT o.id, o.restaurant_id, o.table_id, t.name as table_name, o.table_session_id,
		       o.order_number, o.status, o.subtotal, o.tax, o.service_charge, o.discount, o.total,
		       o.notes, o.created_at, o.updated_at
		FROM orders o
		LEFT JOIN tables t ON t.id = o.table_id
		WHERE o.id = $1
	`
	var o Order
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&o.ID, &o.RestaurantID, &o.TableID, &o.TableName, &o.TableSessionID,
		&o.OrderNumber, &o.Status, &o.Subtotal, &o.Tax, &o.ServiceCharge, &o.Discount, &o.Total,
		&o.Notes, &o.CreatedAt, &o.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrOrderNotFound
		}
		return nil, err
	}

	items, err := r.GetOrderItems(ctx, o.ID)
	if err == nil {
		o.Items = items
	}
	return &o, nil
}

func (r *repository) GetByOrderNumber(ctx context.Context, orderNum string) (*Order, error) {
	query := `
		SELECT o.id, o.restaurant_id, o.table_id, t.name as table_name, o.table_session_id,
		       o.order_number, o.status, o.subtotal, o.tax, o.service_charge, o.discount, o.total,
		       o.notes, o.created_at, o.updated_at
		FROM orders o
		LEFT JOIN tables t ON t.id = o.table_id
		WHERE o.order_number = $1
	`
	var o Order
	err := r.pool.QueryRow(ctx, query, orderNum).Scan(
		&o.ID, &o.RestaurantID, &o.TableID, &o.TableName, &o.TableSessionID,
		&o.OrderNumber, &o.Status, &o.Subtotal, &o.Tax, &o.ServiceCharge, &o.Discount, &o.Total,
		&o.Notes, &o.CreatedAt, &o.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrOrderNotFound
		}
		return nil, err
	}

	items, err := r.GetOrderItems(ctx, o.ID)
	if err == nil {
		o.Items = items
	}
	return &o, nil
}

func (r *repository) List(ctx context.Context, restaurantID string, status *Status) ([]Order, error) {
	query := `
		SELECT o.id, o.restaurant_id, o.table_id, t.name as table_name, o.table_session_id,
		       o.order_number, o.status, o.subtotal, o.tax, o.service_charge, o.discount, o.total,
		       o.notes, o.created_at, o.updated_at
		FROM orders o
		LEFT JOIN tables t ON t.id = o.table_id
		WHERE o.restaurant_id = $1 AND ($2::varchar IS NULL OR o.status = $2)
		ORDER BY o.created_at DESC
		LIMIT 100
	`
	rows, err := r.pool.Query(ctx, query, restaurantID, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []Order
	for rows.Next() {
		var o Order
		if err := rows.Scan(
			&o.ID, &o.RestaurantID, &o.TableID, &o.TableName, &o.TableSessionID,
			&o.OrderNumber, &o.Status, &o.Subtotal, &o.Tax, &o.ServiceCharge, &o.Discount, &o.Total,
			&o.Notes, &o.CreatedAt, &o.UpdatedAt,
		); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}

	// Fetch items for each order
	for i := range orders {
		items, err := r.GetOrderItems(ctx, orders[i].ID)
		if err == nil {
			orders[i].Items = items
		}
	}

	return orders, rows.Err()
}

func (r *repository) ListKitchenOrders(ctx context.Context, restaurantID string) ([]Order, error) {
	query := `
		SELECT o.id, o.restaurant_id, o.table_id, t.name as table_name, o.table_session_id,
		       o.order_number, o.status, o.subtotal, o.tax, o.service_charge, o.discount, o.total,
		       o.notes, o.created_at, o.updated_at
		FROM orders o
		LEFT JOIN tables t ON t.id = o.table_id
		WHERE o.restaurant_id = $1 
		  AND (o.status IN ('CONFIRMED', 'PREPARING', 'READY') 
		       OR (o.status = 'COMPLETED' AND o.updated_at >= NOW() - INTERVAL '12 hours'))
		ORDER BY o.created_at ASC
	`
	rows, err := r.pool.Query(ctx, query, restaurantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []Order
	for rows.Next() {
		var o Order
		if err := rows.Scan(
			&o.ID, &o.RestaurantID, &o.TableID, &o.TableName, &o.TableSessionID,
			&o.OrderNumber, &o.Status, &o.Subtotal, &o.Tax, &o.ServiceCharge, &o.Discount, &o.Total,
			&o.Notes, &o.CreatedAt, &o.UpdatedAt,
		); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}

	for i := range orders {
		items, err := r.GetOrderItems(ctx, orders[i].ID)
		if err == nil {
			orders[i].Items = items
		}
	}

	return orders, rows.Err()
}

func (r *repository) GetOrderItems(ctx context.Context, orderID string) ([]OrderItem, error) {
	query := `
		SELECT id, order_id, menu_id, menu_name_snapshot, unit_price, quantity, subtotal, selected_modifiers, notes, created_at
		FROM order_items
		WHERE order_id = $1
		ORDER BY created_at ASC
	`
	rows, err := r.pool.Query(ctx, query, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []OrderItem
	for rows.Next() {
		var it OrderItem
		var modsJSON []byte
		if err := rows.Scan(
			&it.ID, &it.OrderID, &it.MenuID, &it.MenuNameSnapshot, &it.UnitPrice, &it.Quantity, &it.Subtotal, &modsJSON, &it.Notes, &it.CreatedAt,
		); err != nil {
			return nil, err
		}
		if len(modsJSON) > 0 {
			_ = json.Unmarshal(modsJSON, &it.SelectedModifiers)
		}
		items = append(items, it)
	}
	return items, rows.Err()
}

func (r *repository) UpdateStatus(ctx context.Context, id string, newStatus Status, changedBy *string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var currentStatus Status
	err = tx.QueryRow(ctx, `SELECT status FROM orders WHERE id = $1 FOR UPDATE`, id).Scan(&currentStatus)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrOrderNotFound
		}
		return err
	}

	if !IsValidTransition(currentStatus, newStatus) {
		return ErrInvalidStatusOrder
	}

	now := time.Now()
	_, err = tx.Exec(ctx, `UPDATE orders SET status = $1, updated_at = $2 WHERE id = $3`, newStatus, now, id)
	if err != nil {
		return err
	}

	historyQuery := `
		INSERT INTO order_status_history (id, order_id, from_status, to_status, changed_by, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	hID := "osh_" + id + "_" + string(newStatus)
	_, err = tx.Exec(ctx, historyQuery, hID, id, currentStatus, newStatus, changedBy, now)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *repository) GetTodayStats(ctx context.Context, restaurantID string) (revenue int64, orderCount int, avgOrder int64, paidCount int, cancelledCount int, err error) {
	query := `
		SELECT 
			COALESCE(SUM(CASE WHEN status != 'CANCELLED' AND status != 'WAITING_PAYMENT' THEN total ELSE 0 END), 0) as revenue,
			COUNT(id) as total_orders,
			COUNT(CASE WHEN status != 'CANCELLED' AND status != 'WAITING_PAYMENT' THEN 1 END) as paid_orders,
			COUNT(CASE WHEN status = 'CANCELLED' THEN 1 END) as cancelled_orders
		FROM orders
		WHERE restaurant_id = $1 AND created_at >= CURRENT_DATE
	`
	err = r.pool.QueryRow(ctx, query, restaurantID).Scan(&revenue, &orderCount, &paidCount, &cancelledCount)
	if err != nil {
		return 0, 0, 0, 0, 0, err
	}
	if paidCount > 0 {
		avgOrder = revenue / int64(paidCount)
	}
	return revenue, orderCount, avgOrder, paidCount, cancelledCount, nil
}
