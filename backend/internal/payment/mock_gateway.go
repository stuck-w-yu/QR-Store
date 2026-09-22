package payment

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type MockGateway struct{}

func NewMockGateway() *MockGateway {
	return &MockGateway{}
}

func (m *MockGateway) Name() string {
	return "mock"
}

func (m *MockGateway) CreatePayment(ctx context.Context, req CreatePaymentRequest) (*PaymentResponse, error) {
	txID := "mock_tx_" + uuid.New().String()[:12]
	expiredAt := time.Now().Add(15 * time.Minute)

	// Simulated dummy QRIS data string
	qrString := fmt.Sprintf("00020101021226600016ID.CO.QRSTORE.WWW011893600914%s520458125303360540%d5802ID5913QR-Store6007JAKARTA", req.OrderID, req.Amount)

	return &PaymentResponse{
		TransactionID: txID,
		PaymentURL:    fmt.Sprintf("https://mock-gateway.example.com/pay/%s", txID),
		QRString:      qrString,
		ExpiredAt:     expiredAt,
	}, nil
}

type MockWebhookPayload struct {
	EventID       string `json:"event_id"`
	OrderID       string `json:"order_id"`
	TransactionID string `json:"transaction_id"`
	Amount        int64  `json:"amount"`
	Status        string `json:"status"` // PAID, FAILED, EXPIRED
	PaymentMethod string `json:"payment_method"`
	Signature     string `json:"signature"`
}

func (m *MockGateway) VerifyWebhook(headers http.Header, rawBody []byte) (*WebhookResult, error) {
	var payload MockWebhookPayload
	if err := json.Unmarshal(rawBody, &payload); err != nil {
		return nil, errors.New("invalid webhook payload json")
	}

	// Signature verification simulation
	expectedSig := headers.Get("X-Mock-Signature")
	if expectedSig != "" && expectedSig == "invalid-sig" {
		return nil, errors.New("invalid webhook signature")
	}

	status := StatusPending
	switch payload.Status {
	case "PAID", "SETTLEMENT", "SUCCESS":
		status = StatusPaid
	case "FAILED":
		status = StatusFailed
	case "EXPIRED":
		status = StatusExpired
	case "CANCELLED":
		status = StatusCancelled
	}

	now := time.Now()
	var paidAt *time.Time
	if status == StatusPaid {
		paidAt = &now
	}

	return &WebhookResult{
		EventID:               payload.EventID,
		EventType:             "payment.status_changed",
		OrderID:               payload.OrderID,
		ProviderTransactionID: payload.TransactionID,
		Amount:                payload.Amount,
		Status:                status,
		PaymentMethod:         payload.PaymentMethod,
		PaidAt:                paidAt,
	}, nil
}
