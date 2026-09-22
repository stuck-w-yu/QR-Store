package cashier

import "time"

type ShiftStatus string

const (
	ShiftStatusOpen      ShiftStatus = "OPEN"
	ShiftStatusClosing   ShiftStatus = "CLOSING"
	ShiftStatusClosed    ShiftStatus = "CLOSED"
	ShiftStatusCancelled ShiftStatus = "CANCELLED"
)

func IsValidShiftTransition(from, to ShiftStatus) bool {
	switch from {
	case ShiftStatusOpen:
		return to == ShiftStatusClosing || to == ShiftStatusClosed || to == ShiftStatusCancelled
	case ShiftStatusClosing:
		return to == ShiftStatusClosed || to == ShiftStatusOpen
	case ShiftStatusClosed, ShiftStatusCancelled:
		return false // terminal states
	default:
		return false
	}
}

type TransactionType string

const (
	TxTypeSale       TransactionType = "SALE"
	TxTypeRefund     TransactionType = "REFUND"
	TxTypeVoid       TransactionType = "VOID"
	TxTypeAdjustment TransactionType = "ADJUSTMENT"
)

type RefundStatus string

const (
	RefundStatusPending   RefundStatus = "PENDING"
	RefundStatusCompleted RefundStatus = "COMPLETED"
	RefundStatusRejected  RefundStatus = "REJECTED"
)

type Register struct {
	ID           string    `json:"id"`
	RestaurantID string    `json:"restaurant_id"`
	OutletID     *string   `json:"outlet_id,omitempty"`
	Name         string    `json:"name"`
	Status       string    `json:"status"` // ACTIVE, INACTIVE
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type CashierShift struct {
	ID             string      `json:"id"`
	RestaurantID   string      `json:"restaurant_id"`
	OutletID       *string     `json:"outlet_id,omitempty"`
	RegisterID     string      `json:"register_id"`
	RegisterName   string      `json:"register_name,omitempty"`
	CashierID      string      `json:"cashier_id"`
	CashierName    string      `json:"cashier_name,omitempty"`
	Status         ShiftStatus `json:"status"`
	OpeningBalance int64       `json:"opening_balance"`
	OpenedAt       time.Time   `json:"opened_at"`
	ClosedAt       *time.Time  `json:"closed_at,omitempty"`
	ExpectedTotal  int64       `json:"expected_total"`
	ActualTotal    *int64      `json:"actual_total,omitempty"`
	Difference     *int64      `json:"difference,omitempty"`
	CreatedAt      time.Time   `json:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at"`
}

type CashierShiftTransaction struct {
	ID        string                 `json:"id"`
	ShiftID   string                 `json:"shift_id"`
	OrderID   *string                `json:"order_id,omitempty"`
	PaymentID *string                `json:"payment_id,omitempty"`
	Type      TransactionType        `json:"type"`
	Amount    int64                  `json:"amount"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt time.Time              `json:"created_at"`
}

type ShiftClosing struct {
	ID              string    `json:"id"`
	ShiftID         string    `json:"shift_id"`
	GrossSales      int64     `json:"gross_sales"`
	RefundTotal     int64     `json:"refund_total"`
	VoidTotal       int64     `json:"void_total"`
	AdjustmentTotal int64     `json:"adjustment_total"`
	NetSales        int64     `json:"net_sales"`
	ExpectedAmount  int64     `json:"expected_amount"`
	ActualAmount    *int64    `json:"actual_amount,omitempty"`
	Difference      *int64    `json:"difference,omitempty"`
	Notes           *string   `json:"notes,omitempty"`
	ClosedBy        string    `json:"closed_by"`
	ClosedByName    string    `json:"closed_by_name,omitempty"`
	ClosedAt        time.Time `json:"closed_at"`
}

type Refund struct {
	ID              string       `json:"id"`
	RestaurantID    string       `json:"restaurant_id"`
	OrderID         string       `json:"order_id"`
	OrderNumber     string       `json:"order_number,omitempty"`
	PaymentID       string       `json:"payment_id"`
	Amount          int64        `json:"amount"`
	Reason          string       `json:"reason"`
	Status          RefundStatus `json:"status"`
	RequestedBy     string       `json:"requested_by"`
	RequestedByName string       `json:"requested_by_name,omitempty"`
	ApprovedBy      *string      `json:"approved_by,omitempty"`
	CreatedAt       time.Time    `json:"created_at"`
	CompletedAt     *time.Time   `json:"completed_at,omitempty"`
}

type ShiftSummary struct {
	ShiftID         string           `json:"shift_id"`
	Status          ShiftStatus      `json:"status"`
	RegisterID      string           `json:"register_id"`
	RegisterName    string           `json:"register_name"`
	CashierID       string           `json:"cashier_id"`
	CashierName     string           `json:"cashier_name"`
	OpeningBalance  int64            `json:"opening_balance"`
	OpenedAt        time.Time        `json:"opened_at"`
	ClosedAt        *time.Time       `json:"closed_at,omitempty"`
	OrdersCount     int              `json:"orders_count"`
	GrossSales      int64            `json:"gross_sales"`
	RefundTotal     int64            `json:"refund_total"`
	VoidTotal       int64            `json:"void_total"`
	AdjustmentTotal int64            `json:"adjustment_total"`
	NetSales        int64            `json:"net_sales"`
	ExpectedAmount  int64            `json:"expected_amount"`
	PaymentMethods  map[string]int64 `json:"payment_methods"`
}

// CalculateNetSales: Gross Sales - Refund - Void + Adjustment
func CalculateNetSales(gross, refund, void, adjustment int64) int64 {
	return gross - refund - void + adjustment
}

// CalculateDifference: Actual Amount - Expected Amount
func CalculateDifference(expected, actual int64) int64 {
	return actual - expected
}
