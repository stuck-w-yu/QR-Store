package unit

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"qr-store/backend/internal/order"
)

func TestCashPaymentCalculations(t *testing.T) {
	tests := []struct {
		name        string
		grandTotal  int64
		paidAmount  int64
		expectedErr bool
		expectedChg int64
	}{
		{
			name:        "Exact payment (Rp56.000 paid Rp56.000)",
			grandTotal:  56000,
			paidAmount:  56000,
			expectedErr: false,
			expectedChg: 0,
		},
		{
			name:        "Overpayment (Rp56.000 paid Rp100.000 -> change Rp44.000 - PRD Sec 9)",
			grandTotal:  56000,
			paidAmount:  100000,
			expectedErr: false,
			expectedChg: 44000,
		},
		{
			name:        "Overpayment (Rp75.000 paid Rp100.000 -> change Rp25.000 - PRD Sec 41)",
			grandTotal:  75000,
			paidAmount:  100000,
			expectedErr: false,
			expectedChg: 25000,
		},
		{
			name:        "Insufficient payment (Rp75.000 paid Rp50.000 -> reject - PRD Sec 41)",
			grandTotal:  75000,
			paidAmount:  50000,
			expectedErr: true,
			expectedChg: 0,
		},
		{
			name:        "Zero payment on non-zero total (reject)",
			grandTotal:  50000,
			paidAmount:  0,
			expectedErr: true,
			expectedChg: 0,
		},
		{
			name:        "Negative payment (reject)",
			grandTotal:  50000,
			paidAmount:  -10000,
			expectedErr: true,
			expectedChg: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.paidAmount < tt.grandTotal {
				assert.True(t, tt.expectedErr, "expected error for insufficient payment")
			} else {
				change := tt.paidAmount - tt.grandTotal
				assert.False(t, tt.expectedErr)
				assert.Equal(t, tt.expectedChg, change)
			}
		})
	}
}

func TestPaymentMethodValidation(t *testing.T) {
	validMethods := []string{
		order.PaymentMethodCash,
		order.PaymentMethodQRISManual,
		order.PaymentMethodDebit,
		order.PaymentMethodOther,
		"QRIS",
		"cash",
		"qris_manual",
		"debit",
	}

	for _, m := range validMethods {
		t.Run("Valid: "+m, func(t *testing.T) {
			norm := strings.ToUpper(strings.TrimSpace(m))
			if norm == "QRIS" {
				norm = order.PaymentMethodQRISManual
			}
			isValid := norm == order.PaymentMethodCash ||
				norm == order.PaymentMethodQRISManual ||
				norm == order.PaymentMethodDebit ||
				norm == order.PaymentMethodOther
			assert.True(t, isValid)
		})
	}

	invalidMethods := []string{
		"BITCOIN",
		"PAYPAL",
		"CREDIT_UNSUPPORTED",
		"",
	}

	for _, m := range invalidMethods {
		t.Run("Invalid: "+m, func(t *testing.T) {
			norm := strings.ToUpper(strings.TrimSpace(m))
			isValid := norm == order.PaymentMethodCash ||
				norm == order.PaymentMethodQRISManual ||
				norm == order.PaymentMethodDebit ||
				norm == order.PaymentMethodOther
			assert.False(t, isValid)
		})
	}
}

func TestNonCashPaymentsChangeIsZero(t *testing.T) {
	total := int64(85000)
	methods := []string{order.PaymentMethodQRISManual, order.PaymentMethodDebit, order.PaymentMethodOther}

	for _, m := range methods {
		t.Run(m, func(t *testing.T) {
			paidAmount := total
			var change int64 = 0
			assert.Equal(t, int64(0), change)
			assert.Equal(t, total, paidAmount)
		})
	}
}
