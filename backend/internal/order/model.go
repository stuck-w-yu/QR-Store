package order

import "time"

type Status string

const (
	StatusWaitingPayment Status = "WAITING_PAYMENT"
	StatusConfirmed      Status = "CONFIRMED"
	StatusPreparing      Status = "PREPARING"
	StatusReady          Status = "READY"
	StatusCompleted      Status = "COMPLETED"
	StatusCancelled      Status = "CANCELLED"
)

func IsValidTransition(from, to Status) bool {
	switch from {
	case StatusWaitingPayment:
		return to == StatusConfirmed || to == StatusCancelled
	case StatusConfirmed:
		return to == StatusPreparing || to == StatusCancelled
	case StatusPreparing:
		return to == StatusReady || to == StatusCancelled
	case StatusReady:
		return to == StatusCompleted || to == StatusCancelled
	case StatusCompleted, StatusCancelled:
		return false // terminal states
	default:
		return false
	}
}

type Order struct {
	ID             string      `json:"id"`
	RestaurantID   string      `json:"restaurant_id"`
	TableID        string      `json:"table_id"`
	TableName      string      `json:"table_name,omitempty"`
	TableSessionID *string     `json:"table_session_id,omitempty"`
	OrderNumber    string      `json:"order_number"`
	Status         Status      `json:"status"`
	Subtotal       int64       `json:"subtotal"`
	Tax            int64       `json:"tax"`
	ServiceCharge  int64       `json:"service_charge"`
	Discount       int64       `json:"discount"`
	Total          int64       `json:"total"`
	Notes          *string     `json:"notes"`
	Items          []OrderItem `json:"items,omitempty"`
	CreatedAt      time.Time   `json:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at"`
}

type SelectedModifierOption struct {
	ModifierID      string `json:"modifier_id"`
	ModifierName    string `json:"modifier_name"`
	OptionID        string `json:"option_id"`
	OptionName      string `json:"option_name"`
	AdditionalPrice int64  `json:"additional_price"`
}

type OrderItem struct {
	ID                string                   `json:"id"`
	OrderID           string                   `json:"order_id"`
	MenuID            *string                  `json:"menu_id"`
	MenuNameSnapshot  string                   `json:"menu_name_snapshot"`
	UnitPrice         int64                    `json:"unit_price"`
	Quantity          int                      `json:"quantity"`
	Subtotal          int64                    `json:"subtotal"`
	SelectedModifiers []SelectedModifierOption `json:"selected_modifiers"`
	Notes             *string                  `json:"notes"`
	CreatedAt         time.Time                `json:"created_at"`
}

type StatusHistory struct {
	ID         string    `json:"id"`
	OrderID    string    `json:"order_id"`
	FromStatus *Status   `json:"from_status"`
	ToStatus   Status    `json:"to_status"`
	ChangedBy  *string   `json:"changed_by"`
	CreatedAt  time.Time `json:"created_at"`
}
