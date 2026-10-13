package payment

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
)

// webhookMaxSkew bounds how old a signed webhook may be (anti-replay).
const webhookMaxSkew = 5 * time.Minute

type MockGateway struct {
	WebhookSecret string
}

func NewMockGateway(webhookSecret string) *MockGateway {
	return &MockGateway{WebhookSecret: webhookSecret}
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

// SignSignature produces the HMAC-SHA256 signature expected in the
// X-Mock-Signature header (hex encoded).
func (m *MockGateway) SignSignature(rawBody []byte) string {
	mac := hmac.New(sha256.New, []byte(m.WebhookSecret))
	mac.Write(rawBody)
	return hex.EncodeToString(mac.Sum(nil))
}

func (m *MockGateway) VerifyWebhook(headers http.Header, rawBody []byte) (*WebhookResult, error) {
	if m.WebhookSecret == "" {
		return nil, errors.New("webhook secret is not configured")
	}

	// Anti-replay: require a fresh timestamp within the allowed skew.
	tsHeader := headers.Get("X-Webhook-Timestamp")
	if tsHeader != "" {
		ts, err := strconv.ParseInt(tsHeader, 10, 64)
		if err != nil {
			return nil, errors.New("invalid webhook timestamp")
		}
		if time.Since(time.Unix(ts, 0)).Abs() > webhookMaxSkew {
			return nil, errors.New("webhook timestamp outside allowed window")
		}
	}

	providedSig := headers.Get("X-Mock-Signature")
	if providedSig == "" {
		return nil, errors.New("missing webhook signature")
	}
	expectedSig := m.SignSignature(rawBody)
	if !hmac.Equal([]byte(expectedSig), []byte(providedSig)) {
		return nil, errors.New("invalid webhook signature")
	}

	var payload MockWebhookPayload
	if err := json.Unmarshal(rawBody, &payload); err != nil {
		return nil, errors.New("invalid webhook payload json")
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
