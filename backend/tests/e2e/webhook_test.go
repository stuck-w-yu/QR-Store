//go:build e2e

package e2e

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"qr-store/backend/internal/payment"
)

func TestWebhookEvent_DuplicateDetection(t *testing.T) {
	pool := testDB(t)
	ctx := context.Background()

	repo := payment.NewRepository(pool)
	now := time.Now()
	et := "payment.status_changed"

	first := &payment.WebhookEvent{
		ID: "whe_1", Provider: "mock", EventID: "evt_123",
		EventType: &et, Payload: `{"a":1}`, Processed: false, CreatedAt: now,
	}

	dup, err := repo.RecordWebhookEvent(ctx, first)
	require.NoError(t, err)
	assert.False(t, dup, "first delivery must not be a duplicate")

	// Same (provider, event_id) again => duplicate, and no error.
	second := &payment.WebhookEvent{
		ID: "whe_2", Provider: "mock", EventID: "evt_123",
		EventType: &et, Payload: `{"a":1}`, Processed: false, CreatedAt: now,
	}
	dup, err = repo.RecordWebhookEvent(ctx, second)
	require.NoError(t, err)
	assert.True(t, dup, "second delivery with same event_id must be flagged as duplicate")

	// Different provider with same event_id is a distinct event.
	third := &payment.WebhookEvent{
		ID: "whe_3", Provider: "midtrans", EventID: "evt_123",
		EventType: &et, Payload: `{"a":1}`, Processed: false, CreatedAt: now,
	}
	dup, err = repo.RecordWebhookEvent(ctx, third)
	require.NoError(t, err)
	assert.False(t, dup, "same event_id under a different provider is not a duplicate")
}

func TestWebhookEvent_MarkProcessed(t *testing.T) {
	pool := testDB(t)
	ctx := context.Background()

	repo := payment.NewRepository(pool)
	et := "payment.status_changed"
	ev := &payment.WebhookEvent{
		ID: "whe_mark", Provider: "mock", EventID: "evt_mark",
		EventType: &et, Payload: `{}`, Processed: false, CreatedAt: time.Now(),
	}
	_, err := repo.RecordWebhookEvent(ctx, ev)
	require.NoError(t, err)

	var processed bool
	require.NoError(t, pool.QueryRow(ctx, `SELECT processed FROM payment_webhook_events WHERE id = $1`, "whe_mark").Scan(&processed))
	assert.False(t, processed, "event must start unprocessed")

	require.NoError(t, repo.MarkWebhookEventProcessed(ctx, "whe_mark", true))
	require.NoError(t, pool.QueryRow(ctx, `SELECT processed FROM payment_webhook_events WHERE id = $1`, "whe_mark").Scan(&processed))
	assert.True(t, processed, "event must be marked processed after success")
}
