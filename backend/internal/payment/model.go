package payment

import "time"

type Status string

const (
	StatusPending   Status = "PENDING"
	StatusPaid      Status = "PAID"
	StatusFailed    Status = "FAILED"
	StatusExpired   Status = "EXPIRED"
	StatusCancelled Status = "CANCELLED"
)

type Payment struct {
	ID                    string     `json:"id"`
	RestaurantID          string     `json:"restaurant_id"`
	OrderID               string     `json:"order_id"`
	Provider              string     `json:"provider"`
	ProviderTransactionID *string    `json:"provider_transaction_id"`
	PaymentMethod         *string    `json:"payment_method"`
	Amount                int64      `json:"amount"`
	Status                Status     `json:"status"`
	PaymentURL            *string    `json:"payment_url,omitempty"`
	QRString              *string    `json:"qr_string,omitempty"`
	ExpiredAt             *time.Time `json:"expired_at,omitempty"`
	PaidAt                *time.Time `json:"paid_at,omitempty"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
}

type WebhookEvent struct {
	ID          string     `json:"id"`
	Provider    string     `json:"provider"`
	EventID     string     `json:"event_id"`
	EventType   *string    `json:"event_type"`
	Payload     string     `json:"payload"`
	Processed   bool       `json:"processed"`
	ProcessedAt *time.Time `json:"processed_at"`
	CreatedAt   time.Time  `json:"created_at"`
}
