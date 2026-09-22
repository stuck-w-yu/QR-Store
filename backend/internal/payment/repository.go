package payment

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"qr-store/backend/internal/order"
)

var (
	ErrPaymentNotFound = errors.New("payment not found")
)

type Repository interface {
	Create(ctx context.Context, p *Payment) error
	GetByID(ctx context.Context, id string) (*Payment, error)
	GetByOrderID(ctx context.Context, orderID string) (*Payment, error)
	Update(ctx context.Context, p *Payment) error
	RecordWebhookEvent(ctx context.Context, ev *WebhookEvent) (isDuplicate bool, err error)
	ProcessPaymentSuccess(ctx context.Context, orderID string, paymentID string, paidAt time.Time, method string, txID string) (*order.Order, error)
}

type repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) Repository {
	return &repository{pool: pool}
}

func (r *repository) Create(ctx context.Context, p *Payment) error {
	query := `
		INSERT INTO payments (id, restaurant_id, order_id, provider, provider_transaction_id, payment_method, amount, status, payment_url, qr_string, expired_at, paid_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`
	_, err := r.pool.Exec(ctx, query,
		p.ID, p.RestaurantID, p.OrderID, p.Provider, p.ProviderTransactionID, p.PaymentMethod,
		p.Amount, p.Status, p.PaymentURL, p.QRString, p.ExpiredAt, p.PaidAt, p.CreatedAt, p.UpdatedAt,
	)
	return err
}

func (r *repository) GetByID(ctx context.Context, id string) (*Payment, error) {
	query := `
		SELECT id, restaurant_id, order_id, provider, provider_transaction_id, payment_method, amount, status, payment_url, qr_string, expired_at, paid_at, created_at, updated_at
		FROM payments
		WHERE id = $1
	`
	var p Payment
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&p.ID, &p.RestaurantID, &p.OrderID, &p.Provider, &p.ProviderTransactionID, &p.PaymentMethod,
		&p.Amount, &p.Status, &p.PaymentURL, &p.QRString, &p.ExpiredAt, &p.PaidAt, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrPaymentNotFound
		}
		return nil, err
	}
	return &p, nil
}

func (r *repository) GetByOrderID(ctx context.Context, orderID string) (*Payment, error) {
	query := `
		SELECT id, restaurant_id, order_id, provider, provider_transaction_id, payment_method, amount, status, payment_url, qr_string, expired_at, paid_at, created_at, updated_at
		FROM payments
		WHERE order_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`
	var p Payment
	err := r.pool.QueryRow(ctx, query, orderID).Scan(
		&p.ID, &p.RestaurantID, &p.OrderID, &p.Provider, &p.ProviderTransactionID, &p.PaymentMethod,
		&p.Amount, &p.Status, &p.PaymentURL, &p.QRString, &p.ExpiredAt, &p.PaidAt, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrPaymentNotFound
		}
		return nil, err
	}
	return &p, nil
}

func (r *repository) Update(ctx context.Context, p *Payment) error {
	p.UpdatedAt = time.Now()
	query := `
		UPDATE payments
		SET provider_transaction_id = $1, payment_method = $2, status = $3, paid_at = $4, updated_at = $5
		WHERE id = $6
	`
	_, err := r.pool.Exec(ctx, query,
		p.ProviderTransactionID, p.PaymentMethod, p.Status, p.PaidAt, p.UpdatedAt, p.ID,
	)
	return err
}

func (r *repository) RecordWebhookEvent(ctx context.Context, ev *WebhookEvent) (bool, error) {
	query := `
		INSERT INTO payment_webhook_events (id, provider, event_id, event_type, payload, processed, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (provider, event_id) DO NOTHING
	`
	tag, err := r.pool.Exec(ctx, query,
		ev.ID, ev.Provider, ev.EventID, ev.EventType, ev.Payload, ev.Processed, ev.CreatedAt,
	)
	if err != nil {
		return false, err
	}

	// If RowsAffected is 0, the event was already recorded!
	if tag.RowsAffected() == 0 {
		return true, nil
	}
	return false, nil
}

func (r *repository) ProcessPaymentSuccess(ctx context.Context, orderID string, paymentID string, paidAt time.Time, method string, txID string) (*order.Order, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// 1. Lock and fetch order
	var o order.Order
	orderQuery := `
		SELECT id, restaurant_id, table_id, order_number, status, subtotal, tax, service_charge, discount, total, created_at, updated_at
		FROM orders
		WHERE id = $1
		FOR UPDATE
	`
	err = tx.QueryRow(ctx, orderQuery, orderID).Scan(
		&o.ID, &o.RestaurantID, &o.TableID, &o.OrderNumber, &o.Status,
		&o.Subtotal, &o.Tax, &o.ServiceCharge, &o.Discount, &o.Total, &o.CreatedAt, &o.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	// 2. Update payment status
	now := time.Now()
	paymentUpdate := `
		UPDATE payments
		SET status = 'PAID', paid_at = $1, payment_method = $2, provider_transaction_id = $3, updated_at = $4
		WHERE id = $5
	`
	_, err = tx.Exec(ctx, paymentUpdate, paidAt, method, txID, now, paymentID)
	if err != nil {
		return nil, err
	}

	// 3. Update order status to CONFIRMED
	if o.Status == order.StatusWaitingPayment {
		orderUpdate := `
			UPDATE orders
			SET status = 'CONFIRMED', updated_at = $1
			WHERE id = $2
		`
		_, err = tx.Exec(ctx, orderUpdate, now, orderID)
		if err != nil {
			return nil, err
		}

		historyQuery := `
			INSERT INTO order_status_history (id, order_id, from_status, to_status, changed_by, created_at)
			VALUES ($1, $2, 'WAITING_PAYMENT', 'CONFIRMED', 'WEBHOOK', $3)
		`
		_, err = tx.Exec(ctx, historyQuery, "osh_"+orderID+"_CONFIRMED", orderID, now)
		if err != nil {
			return nil, err
		}
		o.Status = order.StatusConfirmed
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &o, nil
}
