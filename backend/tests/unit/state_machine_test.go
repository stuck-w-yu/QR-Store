package unit

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"qr-store/backend/internal/order"
)

func TestOrderStateMachineTransitions(t *testing.T) {
	tests := []struct {
		name     string
		from     order.Status
		to       order.Status
		expected bool
	}{
		{"Waiting to Confirmed", order.StatusWaitingPayment, order.StatusConfirmed, true},
		{"Waiting to Cancelled", order.StatusWaitingPayment, order.StatusCancelled, true},
		{"Waiting to Preparing (illegal jump)", order.StatusWaitingPayment, order.StatusPreparing, false},
		{"Confirmed to Preparing", order.StatusConfirmed, order.StatusPreparing, true},
		{"Confirmed to Cancelled", order.StatusConfirmed, order.StatusCancelled, true},
		{"Preparing to Ready", order.StatusPreparing, order.StatusReady, true},
		{"Ready to Completed", order.StatusReady, order.StatusCompleted, true},
		{"Completed cannot transition to anything", order.StatusCompleted, order.StatusPreparing, false},
		{"Cancelled cannot transition to anything", order.StatusCancelled, order.StatusConfirmed, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := order.IsValidTransition(tt.from, tt.to)
			assert.Equal(t, tt.expected, valid)
		})
	}
}
