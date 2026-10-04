package cashier

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrRegisterNotFound        = errors.New("register not found")
	ErrRegisterHasActiveShift  = errors.New("cannot delete register: active cashier shift is currently open")
	ErrRegisterHasShiftHistory = errors.New("cannot delete register: register has cashier shift history")
	ErrShiftNotFound           = errors.New("cashier shift not found")
	ErrActiveShiftExists       = errors.New("an active shift already exists for this register or cashier")
	ErrShiftAlreadyClosed      = errors.New("shift is already closed")
	ErrInvalidShiftStatus      = errors.New("invalid shift status transition")
	ErrRefundNotFound          = errors.New("refund not found")
	ErrPendingPaymentsBlocked  = errors.New("closing blocked: pending payments exist")
)

type ShiftFilter struct {
	CashierID  *string
	RegisterID *string
	Status     *ShiftStatus
	StartDate  *time.Time
	EndDate    *time.Time
	Limit      int
	Offset     int
}

type Repository interface {
	// Register
	ListRegisters(ctx context.Context, restaurantID string) ([]Register, error)
	GetRegisterByID(ctx context.Context, id string) (*Register, error)
	CreateRegister(ctx context.Context, reg *Register) error
	UpdateRegister(ctx context.Context, reg *Register) error
	DeleteRegister(ctx context.Context, id string) error

	// Shift
	CreateShift(ctx context.Context, shift *CashierShift) error
	GetShiftByID(ctx context.Context, id string) (*CashierShift, error)
	GetActiveShiftByRegister(ctx context.Context, registerID string) (*CashierShift, error)
	GetActiveShiftByCashier(ctx context.Context, cashierID string) (*CashierShift, error)
	GetActiveShiftByRestaurant(ctx context.Context, restaurantID string) (*CashierShift, error)
	ListShifts(ctx context.Context, restaurantID string, filters ShiftFilter) ([]CashierShift, error)
	CloseShiftTx(ctx context.Context, shift *CashierShift, closing *ShiftClosing) error

	// Transactions
	CreateTransaction(ctx context.Context, tx *CashierShiftTransaction) error
	ListTransactions(ctx context.Context, shiftID string, txType *TransactionType, page, limit int) ([]CashierShiftTransaction, int, error)

	// Closings
	GetClosingByShiftID(ctx context.Context, shiftID string) (*ShiftClosing, error)

	// Refunds
	CreateRefund(ctx context.Context, refund *Refund) error
	GetRefundByID(ctx context.Context, id string) (*Refund, error)
	ListRefundsByOrder(ctx context.Context, orderID string) ([]Refund, error)

	// Audit Logs
	CreateAuditLog(ctx context.Context, restaurantID, userID, action, entityType, entityID string, metadata map[string]interface{}) error

	// Summary Calculation
	GetShiftSummary(ctx context.Context, shiftID string) (*ShiftSummary, error)
	CountPendingPayments(ctx context.Context, restaurantID string, since time.Time) (int, error)
}

type repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) Repository {
	return &repository{pool: pool}
}

// ----------------- REGISTERS -----------------

func (r *repository) ListRegisters(ctx context.Context, restaurantID string) ([]Register, error) {
	query := `
		SELECT id, restaurant_id, outlet_id, name, status, created_at, updated_at
		FROM registers
		WHERE restaurant_id = $1
		ORDER BY created_at ASC
	`
	rows, err := r.pool.Query(ctx, query, restaurantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Register
	for rows.Next() {
		var reg Register
		if err := rows.Scan(&reg.ID, &reg.RestaurantID, &reg.OutletID, &reg.Name, &reg.Status, &reg.CreatedAt, &reg.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, reg)
	}
	return list, nil
}

func (r *repository) GetRegisterByID(ctx context.Context, id string) (*Register, error) {
	query := `
		SELECT id, restaurant_id, outlet_id, name, status, created_at, updated_at
		FROM registers
		WHERE id = $1
	`
	var reg Register
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&reg.ID, &reg.RestaurantID, &reg.OutletID, &reg.Name, &reg.Status, &reg.CreatedAt, &reg.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrRegisterNotFound
		}
		return nil, err
	}
	return &reg, nil
}

func (r *repository) CreateRegister(ctx context.Context, reg *Register) error {
	query := `
		INSERT INTO registers (id, restaurant_id, outlet_id, name, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.pool.Exec(ctx, query,
		reg.ID, reg.RestaurantID, reg.OutletID, reg.Name, reg.Status, reg.CreatedAt, reg.UpdatedAt,
	)
	return err
}

func (r *repository) UpdateRegister(ctx context.Context, reg *Register) error {
	reg.UpdatedAt = time.Now()
	query := `
		UPDATE registers
		SET name = $1, status = $2, updated_at = $3
		WHERE id = $4
	`
	_, err := r.pool.Exec(ctx, query, reg.Name, reg.Status, reg.UpdatedAt, reg.ID)
	return err
}

func (r *repository) DeleteRegister(ctx context.Context, id string) error {
	var exists bool
	err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM registers WHERE id = $1)`, id).Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		return ErrRegisterNotFound
	}

	var hasActiveShift bool
	err = r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM cashier_shifts WHERE register_id = $1 AND status = 'OPEN')`, id).Scan(&hasActiveShift)
	if err != nil {
		return err
	}
	if hasActiveShift {
		return ErrRegisterHasActiveShift
	}

	var hasHistory bool
	err = r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM cashier_shifts WHERE register_id = $1)`, id).Scan(&hasHistory)
	if err != nil {
		return err
	}
	if hasHistory {
		return ErrRegisterHasShiftHistory
	}

	query := `DELETE FROM registers WHERE id = $1`
	_, err = r.pool.Exec(ctx, query, id)
	return err
}

// ----------------- SHIFTS -----------------

func (r *repository) CreateShift(ctx context.Context, shift *CashierShift) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Check if active shift exists on this register
	var countReg int
	err = tx.QueryRow(ctx, `SELECT COUNT(*) FROM cashier_shifts WHERE register_id = $1 AND status = 'OPEN'`, shift.RegisterID).Scan(&countReg)
	if err != nil {
		return err
	}
	if countReg > 0 {
		return ErrActiveShiftExists
	}

	// Check if active shift exists for this cashier
	var countCashier int
	err = tx.QueryRow(ctx, `SELECT COUNT(*) FROM cashier_shifts WHERE cashier_id = $1 AND status = 'OPEN'`, shift.CashierID).Scan(&countCashier)
	if err != nil {
		return err
	}
	if countCashier > 0 {
		return ErrActiveShiftExists
	}

	query := `
		INSERT INTO cashier_shifts (id, restaurant_id, outlet_id, register_id, cashier_id, status, opening_balance, opened_at, expected_total, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err = tx.Exec(ctx, query,
		shift.ID, shift.RestaurantID, shift.OutletID, shift.RegisterID, shift.CashierID,
		shift.Status, shift.OpeningBalance, shift.OpenedAt, shift.ExpectedTotal, shift.CreatedAt, shift.UpdatedAt,
	)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *repository) GetShiftByID(ctx context.Context, id string) (*CashierShift, error) {
	query := `
		SELECT s.id, s.restaurant_id, s.outlet_id, s.register_id, r.name, s.cashier_id, u.name,
		       s.status, s.opening_balance, s.opened_at, s.closed_at, s.expected_total, s.actual_total, s.difference,
		       s.created_at, s.updated_at
		FROM cashier_shifts s
		LEFT JOIN registers r ON s.register_id = r.id
		LEFT JOIN users u ON s.cashier_id = u.id
		WHERE s.id = $1
	`
	var s CashierShift
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&s.ID, &s.RestaurantID, &s.OutletID, &s.RegisterID, &s.RegisterName, &s.CashierID, &s.CashierName,
		&s.Status, &s.OpeningBalance, &s.OpenedAt, &s.ClosedAt, &s.ExpectedTotal, &s.ActualTotal, &s.Difference,
		&s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrShiftNotFound
		}
		return nil, err
	}
	return &s, nil
}

func (r *repository) GetActiveShiftByRegister(ctx context.Context, registerID string) (*CashierShift, error) {
	query := `
		SELECT s.id, s.restaurant_id, s.outlet_id, s.register_id, r.name, s.cashier_id, u.name,
		       s.status, s.opening_balance, s.opened_at, s.closed_at, s.expected_total, s.actual_total, s.difference,
		       s.created_at, s.updated_at
		FROM cashier_shifts s
		LEFT JOIN registers r ON s.register_id = r.id
		LEFT JOIN users u ON s.cashier_id = u.id
		WHERE s.register_id = $1 AND s.status = 'OPEN'
		LIMIT 1
	`
	var s CashierShift
	err := r.pool.QueryRow(ctx, query, registerID).Scan(
		&s.ID, &s.RestaurantID, &s.OutletID, &s.RegisterID, &s.RegisterName, &s.CashierID, &s.CashierName,
		&s.Status, &s.OpeningBalance, &s.OpenedAt, &s.ClosedAt, &s.ExpectedTotal, &s.ActualTotal, &s.Difference,
		&s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

func (r *repository) GetActiveShiftByCashier(ctx context.Context, cashierID string) (*CashierShift, error) {
	query := `
		SELECT s.id, s.restaurant_id, s.outlet_id, s.register_id, r.name, s.cashier_id, u.name,
		       s.status, s.opening_balance, s.opened_at, s.closed_at, s.expected_total, s.actual_total, s.difference,
		       s.created_at, s.updated_at
		FROM cashier_shifts s
		LEFT JOIN registers r ON s.register_id = r.id
		LEFT JOIN users u ON s.cashier_id = u.id
		WHERE s.cashier_id = $1 AND s.status = 'OPEN'
		ORDER BY s.opened_at DESC
		LIMIT 1
	`
	var s CashierShift
	err := r.pool.QueryRow(ctx, query, cashierID).Scan(
		&s.ID, &s.RestaurantID, &s.OutletID, &s.RegisterID, &s.RegisterName, &s.CashierID, &s.CashierName,
		&s.Status, &s.OpeningBalance, &s.OpenedAt, &s.ClosedAt, &s.ExpectedTotal, &s.ActualTotal, &s.Difference,
		&s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

func (r *repository) GetActiveShiftByRestaurant(ctx context.Context, restaurantID string) (*CashierShift, error) {
	query := `
		SELECT s.id, s.restaurant_id, s.outlet_id, s.register_id, r.name, s.cashier_id, u.name,
		       s.status, s.opening_balance, s.opened_at, s.closed_at, s.expected_total, s.actual_total, s.difference,
		       s.created_at, s.updated_at
		FROM cashier_shifts s
		LEFT JOIN registers r ON s.register_id = r.id
		LEFT JOIN users u ON s.cashier_id = u.id
		WHERE s.restaurant_id = $1 AND s.status = 'OPEN'
		ORDER BY s.opened_at DESC
		LIMIT 1
	`
	var s CashierShift
	err := r.pool.QueryRow(ctx, query, restaurantID).Scan(
		&s.ID, &s.RestaurantID, &s.OutletID, &s.RegisterID, &s.RegisterName, &s.CashierID, &s.CashierName,
		&s.Status, &s.OpeningBalance, &s.OpenedAt, &s.ClosedAt, &s.ExpectedTotal, &s.ActualTotal, &s.Difference,
		&s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &s, nil
}

func (r *repository) ListShifts(ctx context.Context, restaurantID string, filters ShiftFilter) ([]CashierShift, error) {
	query := `
		SELECT s.id, s.restaurant_id, s.outlet_id, s.register_id, COALESCE(r.name, 'Register'), s.cashier_id, COALESCE(u.name, 'Kasir'),
		       s.status, s.opening_balance, s.opened_at, s.closed_at, s.expected_total, s.actual_total, s.difference,
		       s.created_at, s.updated_at
		FROM cashier_shifts s
		LEFT JOIN registers r ON s.register_id = r.id
		LEFT JOIN users u ON s.cashier_id = u.id
		WHERE s.restaurant_id = $1
	`
	args := []interface{}{restaurantID}
	idx := 2

	if filters.CashierID != nil && *filters.CashierID != "" {
		query += fmt.Sprintf(" AND s.cashier_id = $%d", idx)
		args = append(args, *filters.CashierID)
		idx++
	}
	if filters.RegisterID != nil && *filters.RegisterID != "" {
		query += fmt.Sprintf(" AND s.register_id = $%d", idx)
		args = append(args, *filters.RegisterID)
		idx++
	}
	if filters.Status != nil && *filters.Status != "" {
		query += fmt.Sprintf(" AND s.status = $%d", idx)
		args = append(args, *filters.Status)
		idx++
	}
	if filters.StartDate != nil {
		query += fmt.Sprintf(" AND s.opened_at >= $%d", idx)
		args = append(args, *filters.StartDate)
		idx++
	}
	if filters.EndDate != nil {
		query += fmt.Sprintf(" AND s.opened_at <= $%d", idx)
		args = append(args, *filters.EndDate)
		idx++
	}

	query += " ORDER BY s.opened_at DESC"
	if filters.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d OFFSET %d", filters.Limit, filters.Offset)
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []CashierShift
	for rows.Next() {
		var s CashierShift
		if err := rows.Scan(
			&s.ID, &s.RestaurantID, &s.OutletID, &s.RegisterID, &s.RegisterName, &s.CashierID, &s.CashierName,
			&s.Status, &s.OpeningBalance, &s.OpenedAt, &s.ClosedAt, &s.ExpectedTotal, &s.ActualTotal, &s.Difference,
			&s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, s)
	}
	return list, nil
}

func (r *repository) CloseShiftTx(ctx context.Context, shift *CashierShift, closing *ShiftClosing) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Lock shift row to ensure idempotency and prevent double-closing
	var currentStatus ShiftStatus
	err = tx.QueryRow(ctx, `SELECT status FROM cashier_shifts WHERE id = $1 FOR UPDATE`, shift.ID).Scan(&currentStatus)
	if err != nil {
		return err
	}
	if currentStatus == ShiftStatusClosed {
		return ErrShiftAlreadyClosed
	}

	now := time.Now()
	shiftUpdate := `
		UPDATE cashier_shifts
		SET status = 'CLOSED', closed_at = $1, expected_total = $2, actual_total = $3, difference = $4, updated_at = $5
		WHERE id = $6
	`
	_, err = tx.Exec(ctx, shiftUpdate, now, shift.ExpectedTotal, shift.ActualTotal, shift.Difference, now, shift.ID)
	if err != nil {
		return err
	}

	closingInsert := `
		INSERT INTO shift_closings (id, shift_id, gross_sales, refund_total, void_total, adjustment_total, net_sales, expected_amount, actual_amount, difference, notes, closed_by, closed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		ON CONFLICT (shift_id) DO NOTHING
	`
	_, err = tx.Exec(ctx, closingInsert,
		closing.ID, closing.ShiftID, closing.GrossSales, closing.RefundTotal, closing.VoidTotal, closing.AdjustmentTotal,
		closing.NetSales, closing.ExpectedAmount, closing.ActualAmount, closing.Difference, closing.Notes, closing.ClosedBy, now,
	)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// ----------------- TRANSACTIONS -----------------

func (r *repository) CreateTransaction(ctx context.Context, tx *CashierShiftTransaction) error {
	metaJSON, _ := json.Marshal(tx.Metadata)
	query := `
		INSERT INTO cashier_shift_transactions (id, shift_id, order_id, payment_id, type, amount, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.pool.Exec(ctx, query,
		tx.ID, tx.ShiftID, tx.OrderID, tx.PaymentID, tx.Type, tx.Amount, string(metaJSON), tx.CreatedAt,
	)
	return err
}

func (r *repository) ListTransactions(ctx context.Context, shiftID string, txType *TransactionType, page, limit int) ([]CashierShiftTransaction, int, error) {
	if limit <= 0 {
		limit = 50
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	countQuery := `SELECT COUNT(*) FROM cashier_shift_transactions WHERE shift_id = $1`
	query := `
		SELECT id, shift_id, order_id, payment_id, type, amount, metadata, created_at
		FROM cashier_shift_transactions
		WHERE shift_id = $1
	`
	args := []interface{}{shiftID}
	countArgs := []interface{}{shiftID}

	if txType != nil && *txType != "" {
		countQuery += " AND type = $2"
		countArgs = append(countArgs, *txType)

		query += " AND type = $2"
		args = append(args, *txType)
	}

	var total int
	if err := r.pool.QueryRow(ctx, countQuery, countArgs...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT %d OFFSET %d", limit, offset)
	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []CashierShiftTransaction
	for rows.Next() {
		var tx CashierShiftTransaction
		var metaJSON []byte
		if err := rows.Scan(&tx.ID, &tx.ShiftID, &tx.OrderID, &tx.PaymentID, &tx.Type, &tx.Amount, &metaJSON, &tx.CreatedAt); err != nil {
			return nil, 0, err
		}
		if len(metaJSON) > 0 {
			_ = json.Unmarshal(metaJSON, &tx.Metadata)
		}
		list = append(list, tx)
	}
	return list, total, nil
}

// ----------------- CLOSINGS -----------------

func (r *repository) GetClosingByShiftID(ctx context.Context, shiftID string) (*ShiftClosing, error) {
	query := `
		SELECT c.id, c.shift_id, c.gross_sales, c.refund_total, c.void_total, c.adjustment_total,
		       c.net_sales, c.expected_amount, c.actual_amount, c.difference, c.notes,
		       c.closed_by, u.name, c.closed_at
		FROM shift_closings c
		LEFT JOIN users u ON c.closed_by = u.id
		WHERE c.shift_id = $1
	`
	var cl ShiftClosing
	err := r.pool.QueryRow(ctx, query, shiftID).Scan(
		&cl.ID, &cl.ShiftID, &cl.GrossSales, &cl.RefundTotal, &cl.VoidTotal, &cl.AdjustmentTotal,
		&cl.NetSales, &cl.ExpectedAmount, &cl.ActualAmount, &cl.Difference, &cl.Notes,
		&cl.ClosedBy, &cl.ClosedByName, &cl.ClosedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &cl, nil
}

// ----------------- REFUNDS -----------------

func (r *repository) CreateRefund(ctx context.Context, refund *Refund) error {
	query := `
		INSERT INTO refunds (id, restaurant_id, order_id, payment_id, amount, reason, status, requested_by, approved_by, created_at, completed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err := r.pool.Exec(ctx, query,
		refund.ID, refund.RestaurantID, refund.OrderID, refund.PaymentID, refund.Amount,
		refund.Reason, refund.Status, refund.RequestedBy, refund.ApprovedBy, refund.CreatedAt, refund.CompletedAt,
	)
	return err
}

func (r *repository) GetRefundByID(ctx context.Context, id string) (*Refund, error) {
	query := `
		SELECT r.id, r.restaurant_id, r.order_id, o.order_number, r.payment_id, r.amount, r.reason,
		       r.status, r.requested_by, u.name, r.approved_by, r.created_at, r.completed_at
		FROM refunds r
		LEFT JOIN orders o ON r.order_id = o.id
		LEFT JOIN users u ON r.requested_by = u.id
		WHERE r.id = $1
	`
	var ref Refund
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&ref.ID, &ref.RestaurantID, &ref.OrderID, &ref.OrderNumber, &ref.PaymentID, &ref.Amount, &ref.Reason,
		&ref.Status, &ref.RequestedBy, &ref.RequestedByName, &ref.ApprovedBy, &ref.CreatedAt, &ref.CompletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrRefundNotFound
		}
		return nil, err
	}
	return &ref, nil
}

func (r *repository) ListRefundsByOrder(ctx context.Context, orderID string) ([]Refund, error) {
	query := `
		SELECT r.id, r.restaurant_id, r.order_id, o.order_number, r.payment_id, r.amount, r.reason,
		       r.status, r.requested_by, u.name, r.approved_by, r.created_at, r.completed_at
		FROM refunds r
		LEFT JOIN orders o ON r.order_id = o.id
		LEFT JOIN users u ON r.requested_by = u.id
		WHERE r.order_id = $1
		ORDER BY r.created_at DESC
	`
	rows, err := r.pool.Query(ctx, query, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Refund
	for rows.Next() {
		var ref Refund
		if err := rows.Scan(
			&ref.ID, &ref.RestaurantID, &ref.OrderID, &ref.OrderNumber, &ref.PaymentID, &ref.Amount, &ref.Reason,
			&ref.Status, &ref.RequestedBy, &ref.RequestedByName, &ref.ApprovedBy, &ref.CreatedAt, &ref.CompletedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, ref)
	}
	return list, nil
}

// ----------------- AUDIT LOGS -----------------

func (r *repository) CreateAuditLog(ctx context.Context, restaurantID, userID, action, entityType, entityID string, metadata map[string]interface{}) error {
	metaJSON, _ := json.Marshal(metadata)
	query := `
		INSERT INTO audit_logs (id, restaurant_id, user_id, action, entity_type, entity_id, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	logID := fmt.Sprintf("aud_%d", time.Now().UnixNano())
	_, err := r.pool.Exec(ctx, query, logID, restaurantID, userID, action, entityType, entityID, string(metaJSON), time.Now())
	return err
}

// ----------------- SUMMARY CALCULATION -----------------

func (r *repository) GetShiftSummary(ctx context.Context, shiftID string) (*ShiftSummary, error) {
	shift, err := r.GetShiftByID(ctx, shiftID)
	if err != nil {
		return nil, err
	}

	summary := &ShiftSummary{
		ShiftID:        shift.ID,
		Status:         shift.Status,
		RegisterID:     shift.RegisterID,
		RegisterName:   shift.RegisterName,
		CashierID:      shift.CashierID,
		CashierName:    shift.CashierName,
		OpeningBalance: shift.OpeningBalance,
		OpenedAt:       shift.OpenedAt,
		ClosedAt:       shift.ClosedAt,
		PaymentMethods: make(map[string]int64),
	}

	// 1. Calculate aggregated transactions from cashier_shift_transactions
	queryAgg := `
		SELECT
			COALESCE(SUM(CASE WHEN type = 'SALE' THEN amount ELSE 0 END), 0) AS gross_sales,
			COALESCE(SUM(CASE WHEN type = 'REFUND' THEN amount ELSE 0 END), 0) AS refund_total,
			COALESCE(SUM(CASE WHEN type = 'VOID' THEN amount ELSE 0 END), 0) AS void_total,
			COALESCE(SUM(CASE WHEN type = 'ADJUSTMENT' THEN amount ELSE 0 END), 0) AS adjustment_total,
			COUNT(DISTINCT order_id) AS orders_count
		FROM cashier_shift_transactions
		WHERE shift_id = $1
	`
	err = r.pool.QueryRow(ctx, queryAgg, shiftID).Scan(
		&summary.GrossSales, &summary.RefundTotal, &summary.VoidTotal, &summary.AdjustmentTotal, &summary.OrdersCount,
	)
	if err != nil {
		return nil, err
	}

	summary.NetSales = CalculateNetSales(summary.GrossSales, summary.RefundTotal, summary.VoidTotal, summary.AdjustmentTotal)
	summary.ExpectedAmount = summary.NetSales + summary.OpeningBalance

	// 2. Breakdown by payment method from linked payments
	queryMethod := `
		SELECT COALESCE(p.payment_method, 'QRIS'), COALESCE(SUM(t.amount), 0)
		FROM cashier_shift_transactions t
		JOIN payments p ON t.payment_id = p.id
		WHERE t.shift_id = $1 AND t.type = 'SALE'
		GROUP BY p.payment_method
	`
	rows, err := r.pool.Query(ctx, queryMethod, shiftID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var method string
			var sum int64
			if err := rows.Scan(&method, &sum); err == nil {
				summary.PaymentMethods[method] = sum
			}
		}
	}

	return summary, nil
}

func (r *repository) CountPendingPayments(ctx context.Context, restaurantID string, since time.Time) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM payments
		WHERE restaurant_id = $1 AND status = 'PENDING' AND created_at >= $2
	`
	var count int
	err := r.pool.QueryRow(ctx, query, restaurantID, since).Scan(&count)
	return count, err
}
