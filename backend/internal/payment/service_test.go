package payment

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"qr-store/backend/internal/order"
	"qr-store/backend/pkg/logger"
)

// TestMain initializes the global logger so services under test can log safely.
func TestMain(m *testing.M) {
	logger.Init("production")
	os.Exit(m.Run())
}

// ---- fakes ----

type fakePaymentRepo struct {
	events        map[string]bool // eventID -> exists
	processed     map[string]bool // event id -> processed flag
	orderPaidCall int
	successOrder  *order.Order
	webhookOrder  string
}

func newFakePaymentRepo() *fakePaymentRepo {
	return &fakePaymentRepo{
		events:    map[string]bool{},
		processed: map[string]bool{},
	}
}

func (f *fakePaymentRepo) Create(ctx context.Context, p *Payment) error { return nil }
func (f *fakePaymentRepo) GetByID(ctx context.Context, id string) (*Payment, error) {
	return &Payment{ID: id}, nil
}
func (f *fakePaymentRepo) GetByOrderID(ctx context.Context, orderID string) (*Payment, error) {
	return &Payment{ID: "pay_test", OrderID: orderID}, nil
}
func (f *fakePaymentRepo) Update(ctx context.Context, p *Payment) error { return nil }

func (f *fakePaymentRepo) RecordWebhookEvent(ctx context.Context, ev *WebhookEvent) (bool, error) {
	key := ev.Provider + ":" + ev.EventID
	if f.events[key] {
		return true, nil
	}
	f.events[key] = true
	f.processed[ev.ID] = false
	return false, nil
}

func (f *fakePaymentRepo) MarkWebhookEventProcessed(ctx context.Context, id string, processed bool) error {
	f.processed[id] = processed
	return nil
}

func (f *fakePaymentRepo) ProcessPaymentSuccess(ctx context.Context, orderID, paymentID string, paidAt time.Time, method, txID string) (*order.Order, error) {
	f.orderPaidCall++
	f.webhookOrder = orderID
	return f.successOrder, nil
}

type fakeOrderRepo struct {
	orders map[string]*order.Order
}

func (f *fakeOrderRepo) Create(ctx context.Context, o *order.Order, items []order.OrderItem) error {
	return nil
}
func (f *fakeOrderRepo) GetByID(ctx context.Context, id string) (*order.Order, error) {
	o, ok := f.orders[id]
	if !ok {
		return nil, order.ErrOrderNotFound
	}
	return o, nil
}
func (f *fakeOrderRepo) GetByOrderNumber(ctx context.Context, n string) (*order.Order, error) {
	return nil, order.ErrOrderNotFound
}
func (f *fakeOrderRepo) List(ctx context.Context, restaurantID string, status *order.Status, limit, offset int) ([]order.Order, error) {
	return nil, nil
}
func (f *fakeOrderRepo) ListKitchenOrders(ctx context.Context, restaurantID string) ([]order.Order, error) {
	return nil, nil
}
func (f *fakeOrderRepo) UpdateStatus(ctx context.Context, id string, s order.Status, by *string) error {
	return nil
}
func (f *fakeOrderRepo) UpdateProofURL(ctx context.Context, id, url string) error { return nil }
func (f *fakeOrderRepo) Delete(ctx context.Context, id, restaurantID string) error {
	return nil
}
func (f *fakeOrderRepo) UpdateDetails(ctx context.Context, id string, notes *string, s *order.Status, ps *order.PaymentStatus) error {
	return nil
}
func (f *fakeOrderRepo) GetOrderItems(ctx context.Context, orderID string) ([]order.OrderItem, error) {
	return nil, nil
}
func (f *fakeOrderRepo) GetTodayStats(ctx context.Context, restaurantID string) (int64, int, int64, int, int, error) {
	return 0, 0, 0, 0, 0, nil
}

// ---- helpers ----

// signedPaidWebhook builds a raw body + headers with a valid HMAC signature,
// simulating a genuine provider callback.
func signedPaidWebhook(t *testing.T, gw *MockGateway, eventID, orderID string, amount int64) (http.Header, []byte) {
	t.Helper()
	payload := MockWebhookPayload{
		EventID:       eventID,
		OrderID:       orderID,
		TransactionID: "tx_" + eventID,
		Amount:        amount,
		Status:        "PAID",
		PaymentMethod: "QRIS",
	}
	body, err := json.Marshal(payload)
	require.NoError(t, err)

	h := make(http.Header)
	h.Set("X-Mock-Signature", gw.SignSignature(body))
	h.Set("X-Webhook-Timestamp", strconv.FormatInt(time.Now().Unix(), 10))
	return h, body
}

func newServiceWithFakes(gw *MockGateway, pr *fakePaymentRepo, or *fakeOrderRepo) Service {
	return NewService(pr, or, gw, nil)
}

// ---- tests ----

func TestProcessWebhook_IdempotentWhenDuplicate(t *testing.T) {
	const secret = "test-webhook-secret-value"
	gw := NewMockGateway(secret)

	pr := newFakePaymentRepo()
	or := &fakeOrderRepo{orders: map[string]*order.Order{
		"ord_1": {ID: "ord_1", RestaurantID: "rst_1", Total: 100000, Status: order.StatusWaitingPayment},
	}}
	pr.successOrder = &order.Order{ID: "ord_1", RestaurantID: "rst_1", Status: order.StatusConfirmed}

	svc := newServiceWithFakes(gw, pr, or)

	hdrs, body := signedPaidWebhook(t, gw, "evt_same", "ord_1", 100000)

	// First delivery processes the order exactly once.
	res, err := svc.ProcessWebhook(context.Background(), hdrs, body)
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, 1, pr.orderPaidCall, "order must be settled exactly once")
	assert.Eventually(t, func() bool { return len(pr.processed) == 1 }, time.Second, 10*time.Millisecond)

	// Duplicate delivery with the SAME event id must be a no-op.
	_, err = svc.ProcessWebhook(context.Background(), hdrs, body)
	require.NoError(t, err)
	assert.Equal(t, 1, pr.orderPaidCall, "duplicate webhook must not settle the order again")
}

func TestProcessWebhook_RejectsTamperedSignature(t *testing.T) {
	gw := NewMockGateway("test-webhook-secret-value")
	pr := newFakePaymentRepo()
	or := &fakeOrderRepo{orders: map[string]*order.Order{
		"ord_1": {ID: "ord_1", RestaurantID: "rst_1", Total: 100000, Status: order.StatusWaitingPayment},
	}}
	svc := newServiceWithFakes(gw, pr, or)

	_, body := signedPaidWebhook(t, gw, "evt_tamper", "ord_1", 100000)
	h := make(http.Header)
	h.Set("X-Mock-Signature", "deadbeef")
	h.Set("X-Webhook-Timestamp", strconv.FormatInt(time.Now().Unix(), 10))

	_, err := svc.ProcessWebhook(context.Background(), h, body)
	require.Error(t, err)
	assert.Equal(t, 0, pr.orderPaidCall, "tampered webhook must never settle the order")
}

func TestProcessWebhook_RejectsAmountMismatch(t *testing.T) {
	gw := NewMockGateway("test-webhook-secret-value")
	pr := newFakePaymentRepo()
	or := &fakeOrderRepo{orders: map[string]*order.Order{
		"ord_1": {ID: "ord_1", RestaurantID: "rst_1", Total: 100000, Status: order.StatusWaitingPayment},
	}}
	svc := newServiceWithFakes(gw, pr, or)

	// Genuine signature but wrong amount.
	hdrs, body := signedPaidWebhook(t, gw, "evt_amt", "ord_1", 5000)
	_, err := svc.ProcessWebhook(context.Background(), hdrs, body)
	require.ErrorIs(t, err, ErrInvalidAmount)
	assert.Equal(t, 0, pr.orderPaidCall)
}

func TestProcessWebhook_UnknownOrderReturnsError(t *testing.T) {
	gw := NewMockGateway("test-webhook-secret-value")
	pr := newFakePaymentRepo()
	or := &fakeOrderRepo{orders: map[string]*order.Order{}}
	svc := newServiceWithFakes(gw, pr, or)

	hdrs, body := signedPaidWebhook(t, gw, "evt_missing", "ord_missing", 1000)
	_, err := svc.ProcessWebhook(context.Background(), hdrs, body)
	require.Error(t, err)
	assert.True(t, errors.Is(err, order.ErrOrderNotFound), "expected wrapped ErrOrderNotFound, got %v", err)
}
