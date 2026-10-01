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
		// Draft transitions
		{"Draft to PendingConfirmation", order.StatusDraft, order.StatusPendingConfirmation, true},
		{"Draft to WaitingPayment", order.StatusDraft, order.StatusWaitingPayment, true},
		{"Draft to Cancelled", order.StatusDraft, order.StatusCancelled, true},
		{"Draft to Ready (illegal)", order.StatusDraft, order.StatusReady, false},

		// Pending Confirmation / Waiting Payment transitions
		{"PendingConfirmation to Confirmed", order.StatusPendingConfirmation, order.StatusConfirmed, true},
		{"PendingConfirmation to Cancelled", order.StatusPendingConfirmation, order.StatusCancelled, true},
		{"PendingConfirmation to Preparing (illegal jump)", order.StatusPendingConfirmation, order.StatusPreparing, false},
		{"Waiting to Confirmed", order.StatusWaitingPayment, order.StatusConfirmed, true},
		{"Waiting to Cancelled", order.StatusWaitingPayment, order.StatusCancelled, true},
		{"Waiting to Preparing (illegal jump)", order.StatusWaitingPayment, order.StatusPreparing, false},

		// Confirmed transitions
		{"Confirmed to Preparing", order.StatusConfirmed, order.StatusPreparing, true},
		{"Confirmed to Cancelled", order.StatusConfirmed, order.StatusCancelled, true},
		{"Confirmed to Ready (illegal jump)", order.StatusConfirmed, order.StatusReady, false},

		// Preparing transitions
		{"Preparing to Ready", order.StatusPreparing, order.StatusReady, true},
		{"Preparing to Cancelled", order.StatusPreparing, order.StatusCancelled, true},
		{"Preparing to Completed (illegal jump)", order.StatusPreparing, order.StatusCompleted, false},

		// Ready transitions
		{"Ready to Served", order.StatusReady, order.StatusServed, true},
		{"Ready to Completed", order.StatusReady, order.StatusCompleted, true},
		{"Ready to Cancelled", order.StatusReady, order.StatusCancelled, true},

		// Served transitions
		{"Served to Completed", order.StatusServed, order.StatusCompleted, true},
		{"Served to Cancelled", order.StatusServed, order.StatusCancelled, true},
		{"Served to Preparing (illegal jump)", order.StatusServed, order.StatusPreparing, false},

		// Terminal states
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
