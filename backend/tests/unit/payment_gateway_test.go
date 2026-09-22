package unit

import (
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"qr-store/backend/internal/payment"
)

func TestMidtransSignatureVerification(t *testing.T) {
	serverKey := "SB-Mid-server-TESTKEY123"
	gateway := payment.NewMidtransGateway(serverKey, true)

	orderID := "ord_test_123"
	statusCode := "200"
	grossAmount := "50000"

	// Correct signature: SHA512(order_id + status_code + gross_amount + server_key)
	hash := sha512.Sum512([]byte(orderID + statusCode + grossAmount + serverKey))
	correctSig := hex.EncodeToString(hash[:])

	payload := payment.MidtransWebhookPayload{
		OrderID:           orderID,
		StatusCode:        statusCode,
		GrossAmount:       grossAmount,
		SignatureKey:      correctSig,
		TransactionStatus: "settlement",
		PaymentType:       "qris",
		TransactionID:     "tx_mid_999",
	}
	body, _ := json.Marshal(payload)

	headers := make(http.Header)
	res, err := gateway.VerifyWebhook(headers, body)
	assert.NoError(t, err)
	assert.Equal(t, payment.StatusPaid, res.Status)
	assert.Equal(t, int64(50000), res.Amount)
	assert.Equal(t, "qris", res.PaymentMethod)

	// Tampered signature
	payload.SignatureKey = "tampered_signature"
	tamperedBody, _ := json.Marshal(payload)
	_, err = gateway.VerifyWebhook(headers, tamperedBody)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid midtrans webhook signature")
}

func TestMockGatewayVerify(t *testing.T) {
	gateway := payment.NewMockGateway()

	payload := payment.MockWebhookPayload{
		EventID:       "ev_001",
		OrderID:       "ord_123",
		TransactionID: "tx_mock_1",
		Amount:        25000,
		Status:        "PAID",
		PaymentMethod: "QRIS",
		Signature:     "valid",
	}
	body, _ := json.Marshal(payload)

	headers := make(http.Header)
	res, err := gateway.VerifyWebhook(headers, body)
	assert.NoError(t, err)
	assert.Equal(t, payment.StatusPaid, res.Status)
	assert.Equal(t, int64(25000), res.Amount)
}
