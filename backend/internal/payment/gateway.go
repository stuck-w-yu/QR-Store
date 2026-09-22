package payment

import (
	"context"
	"net/http"
	"time"
)

type CreatePaymentRequest struct {
	OrderID     string
	OrderNumber string
	Amount      int64
	ItemName    string
}

type PaymentResponse struct {
	TransactionID string
	PaymentURL    string
	QRString      string
	ExpiredAt     time.Time
}

type WebhookResult struct {
	EventID               string
	EventType             string
	OrderID               string
	ProviderTransactionID string
	Amount                int64
	Status                Status
	PaymentMethod         string
	PaidAt                *time.Time
}

type PaymentGateway interface {
	Name() string
	CreatePayment(ctx context.Context, req CreatePaymentRequest) (*PaymentResponse, error)
	VerifyWebhook(headers http.Header, rawBody []byte) (*WebhookResult, error)
}
