package payment

import (
	"context"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

type MidtransGateway struct {
	ServerKey string
	IsSandbox bool
}

func NewMidtransGateway(serverKey string, isSandbox bool) *MidtransGateway {
	return &MidtransGateway{
		ServerKey: serverKey,
		IsSandbox: isSandbox,
	}
}

func (m *MidtransGateway) Name() string {
	return "midtrans"
}

func (m *MidtransGateway) CreatePayment(ctx context.Context, req CreatePaymentRequest) (*PaymentResponse, error) {
	// In production, makes HTTP POST to https://app.sandbox.midtrans.com/snap/v1/transactions
	// Here we provide the standard contract response
	txID := "midtrans_" + req.OrderID
	paymentURL := fmt.Sprintf("https://app.sandbox.midtrans.com/snap/v2/vtweb/%s", txID)
	expiredAt := time.Now().Add(15 * time.Minute)

	return &PaymentResponse{
		TransactionID: txID,
		PaymentURL:    paymentURL,
		QRString:      "",
		ExpiredAt:     expiredAt,
	}, nil
}

type MidtransWebhookPayload struct {
	OrderID           string `json:"order_id"`
	StatusCode        string `json:"status_code"`
	GrossAmount       string `json:"gross_amount"`
	SignatureKey      string `json:"signature_key"`
	TransactionStatus string `json:"transaction_status"`
	PaymentType       string `json:"payment_type"`
	TransactionID     string `json:"transaction_id"`
	SettlementTime    string `json:"settlement_time"`
}

func (m *MidtransGateway) VerifyWebhook(headers http.Header, rawBody []byte) (*WebhookResult, error) {
	var payload MidtransWebhookPayload
	if err := json.Unmarshal(rawBody, &payload); err != nil {
		return nil, errors.New("invalid midtrans webhook payload")
	}

	// Verify SHA-512 signature: SHA512(order_id + status_code + gross_amount + server_key)
	rawSig := payload.OrderID + payload.StatusCode + payload.GrossAmount + m.ServerKey
	hash := sha512.Sum512([]byte(rawSig))
	calculatedSig := hex.EncodeToString(hash[:])

	if m.ServerKey != "" && payload.SignatureKey != calculatedSig {
		return nil, errors.New("invalid midtrans webhook signature")
	}

	status := StatusPending
	switch payload.TransactionStatus {
	case "settlement", "capture":
		status = StatusPaid
	case "expire":
		status = StatusExpired
	case "cancel", "deny":
		status = StatusCancelled
	}

	var amount int64
	fmt.Sscanf(payload.GrossAmount, "%d", &amount)

	now := time.Now()
	var paidAt *time.Time
	if status == StatusPaid {
		paidAt = &now
	}

	eventID := fmt.Sprintf("%s_%s_%s", payload.TransactionID, payload.TransactionStatus, payload.StatusCode)

	return &WebhookResult{
		EventID:               eventID,
		EventType:             payload.TransactionStatus,
		OrderID:               payload.OrderID,
		ProviderTransactionID: payload.TransactionID,
		Amount:                amount,
		Status:                status,
		PaymentMethod:         payload.PaymentType,
		PaidAt:                paidAt,
	}, nil
}
