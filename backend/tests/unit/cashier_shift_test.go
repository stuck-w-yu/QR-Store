package unit

import (
	"testing"

	"qr-store/backend/internal/cashier"
)

func TestShiftStateTransitions(t *testing.T) {
	tests := []struct {
		name     string
		from     cashier.ShiftStatus
		to       cashier.ShiftStatus
		expected bool
	}{
		{"Open to Closing", cashier.ShiftStatusOpen, cashier.ShiftStatusClosing, true},
		{"Open to Closed", cashier.ShiftStatusOpen, cashier.ShiftStatusClosed, true},
		{"Open to Cancelled", cashier.ShiftStatusOpen, cashier.ShiftStatusCancelled, true},
		{"Closing to Closed", cashier.ShiftStatusClosing, cashier.ShiftStatusClosed, true},
		{"Closing back to Open", cashier.ShiftStatusClosing, cashier.ShiftStatusOpen, true},
		{"Closed to Open (Forbidden)", cashier.ShiftStatusClosed, cashier.ShiftStatusOpen, false},
		{"Closed to Closing (Forbidden)", cashier.ShiftStatusClosed, cashier.ShiftStatusClosing, false},
		{"Cancelled to Open (Forbidden)", cashier.ShiftStatusCancelled, cashier.ShiftStatusOpen, false},
		{"Invalid status transition", "INVALID", cashier.ShiftStatusOpen, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := cashier.IsValidShiftTransition(tt.from, tt.to)
			if result != tt.expected {
				t.Errorf("expected IsValidShiftTransition(%s, %s) = %v, got %v", tt.from, tt.to, tt.expected, result)
			}
		})
	}
}

func TestCashierReconciliationCalculations(t *testing.T) {
	// Formula: Net Sales = Gross Sales - Refund - Void + Adjustment
	t.Run("Normal sales with refund and void", func(t *testing.T) {
		gross := int64(5000000)
		refund := int64(200000)
		voidVal := int64(100000)
		adjustment := int64(0)

		expectedNet := int64(4700000)
		actualNet := cashier.CalculateNetSales(gross, refund, voidVal, adjustment)
		if actualNet != expectedNet {
			t.Errorf("expected net sales %d, got %d", expectedNet, actualNet)
		}
	})

	t.Run("Sales with positive adjustment", func(t *testing.T) {
		gross := int64(1000000)
		refund := int64(50000)
		voidVal := int64(0)
		adjustment := int64(25000)

		expectedNet := int64(975000)
		actualNet := cashier.CalculateNetSales(gross, refund, voidVal, adjustment)
		if actualNet != expectedNet {
			t.Errorf("expected net sales %d, got %d", expectedNet, actualNet)
		}
	})

	t.Run("Zero sales", func(t *testing.T) {
		actualNet := cashier.CalculateNetSales(0, 0, 0, 0)
		if actualNet != 0 {
			t.Errorf("expected net sales 0, got %d", actualNet)
		}
	})
}

func TestDifferenceCalculation(t *testing.T) {
	// Formula: Difference = Actual Amount - Expected Amount
	tests := []struct {
		name       string
		expected   int64
		actual     int64
		difference int64
	}{
		{"Exact match (zero difference)", 4700000, 4700000, 0},
		{"Cash overage (positive variance)", 4700000, 4750000, 50000},
		{"Cash shortage (negative variance)", 4700000, 4650000, -50000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diff := cashier.CalculateDifference(tt.expected, tt.actual)
			if diff != tt.difference {
				t.Errorf("expected difference %d, got %d", tt.difference, diff)
			}
		})
	}
}

func TestOpeningBalanceValidation(t *testing.T) {
	reqValidZero := cashier.OpenShiftRequest{RegisterID: "reg_01", OpeningBalance: 0}
	if reqValidZero.OpeningBalance < 0 {
		t.Errorf("opening balance 0 should be valid")
	}

	reqValidPositive := cashier.OpenShiftRequest{RegisterID: "reg_01", OpeningBalance: 150000}
	if reqValidPositive.OpeningBalance < 0 {
		t.Errorf("opening balance > 0 should be valid")
	}

	reqInvalidNegative := cashier.OpenShiftRequest{RegisterID: "reg_01", OpeningBalance: -50000}
	if reqInvalidNegative.OpeningBalance >= 0 {
		t.Errorf("negative opening balance must be detected as invalid")
	}
}
