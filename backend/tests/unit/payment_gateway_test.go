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
	gateway := payment.NewMidtransGateway(serverKey, "whsec_test", true)

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
	gateway := payment.NewMockGateway("test-webhook-secret")

	payload := payment.MockWebhookPayload{
		EventID:       "ev_001",
		OrderID:       "ord_123",
		TransactionID: "tx_mock_1",
		Amount:        25000,
		Status:        "PAID",
		PaymentMethod: "QRIS",
	}
	body, _ := json.Marshal(payload)

	// Missing signature must be rejected.
	headers := make(http.Header)
	_, err := gateway.VerifyWebhook(headers, body)
	assert.Error(t, err)

	// Correct HMAC signature is accepted.
	headers.Set("X-Mock-Signature", gateway.SignSignature(body))
	res, err := gateway.VerifyWebhook(headers, body)
	assert.NoError(t, err)
	assert.Equal(t, payment.StatusPaid, res.Status)
	assert.Equal(t, int64(25000), res.Amount)

	// Tampered signature is rejected.
	headers.Set("X-Mock-Signature", "deadbeef")
	_, err = gateway.VerifyWebhook(headers, body)
	assert.Error(t, err)

	// Stale timestamp is rejected (anti-replay).
	freshHeaders := make(http.Header)
	freshHeaders.Set("X-Mock-Signature", gateway.SignSignature(body))
	freshHeaders.Set("X-Webhook-Timestamp", "1000000") // year 1970
	_, err = gateway.VerifyWebhook(freshHeaders, body)
	assert.Error(t, err)
}
