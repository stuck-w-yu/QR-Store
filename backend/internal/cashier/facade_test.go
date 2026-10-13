package cashier

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"qr-store/backend/internal/order"
	"qr-store/backend/internal/payment"
)

// ---- fakes ----

type fakeOrderService struct {
	lastListRestaurantID string
	lastListLimit        int
	lastListOffset       int
	lastGetID            string
	lastDeleteOrderID    string
	lastDeleteRestoID    string
	lastCancelID         string
	lastCancelStatus     order.Status
}

func (f *fakeOrderService) CreatePublicOrder(ctx context.Context, req order.CreateOrderRequest) (*order.Order, error) {
	return nil, nil
}
func (f *fakeOrderService) GetOrderByID(ctx context.Context, id string) (*order.Order, error) {
	f.lastGetID = id
	return &order.Order{ID: id}, nil
}
func (f *fakeOrderService) GetPublicOrder(ctx context.Context, id, qrToken string) (*order.Order, error) {
	return nil, nil
}
func (f *fakeOrderService) ListOrders(ctx context.Context, restaurantID string, status *order.Status, limit, offset int) ([]order.Order, error) {
	f.lastListRestaurantID = restaurantID
	f.lastListLimit = limit
	f.lastListOffset = offset
	return []order.Order{}, nil
}
func (f *fakeOrderService) ListKitchenOrders(ctx context.Context, restaurantID string) ([]order.Order, error) {
	return nil, nil
}
func (f *fakeOrderService) UpdateOrderStatus(ctx context.Context, id string, newStatus order.Status, changedBy string) error {
	f.lastCancelID = id
	f.lastCancelStatus = newStatus
	return nil
}
func (f *fakeOrderService) SubmitProofOfPayment(ctx context.Context, orderID, qrToken, proofURL string) (*order.Order, error) {
	return nil, nil
}
func (f *fakeOrderService) DeleteOrder(ctx context.Context, id, restaurantID string) error {
	f.lastDeleteOrderID = id
	f.lastDeleteRestoID = restaurantID
	return nil
}
func (f *fakeOrderService) UpdateOrderDetails(ctx context.Context, id, restaurantID string, notes *string, status *order.Status, ps *order.PaymentStatus) (*order.Order, error) {
	return nil, nil
}
func (f *fakeOrderService) GetTodayAnalytics(ctx context.Context, restaurantID string) (map[string]interface{}, error) {
	return nil, nil
}

type fakePaymentService struct {
	lastConfirmOrderID string
	lastProcessReq     payment.ProcessManualPaymentRequest
}

func (f *fakePaymentService) CreatePaymentForOrder(ctx context.Context, orderID string) (*payment.Payment, error) {
	return nil, nil
}
func (f *fakePaymentService) GetPaymentByID(ctx context.Context, id string) (*payment.Payment, error) {
	return nil, nil
}
func (f *fakePaymentService) GetPaymentByOrderID(ctx context.Context, orderID string) (*payment.Payment, error) {
	return nil, nil
}
func (f *fakePaymentService) ProcessWebhook(ctx context.Context, headers http.Header, rawBody []byte) (*payment.WebhookResult, error) {
	return nil, nil
}
func (f *fakePaymentService) ConfirmOrder(ctx context.Context, orderID, confirmedBy string) (*order.Order, error) {
	f.lastConfirmOrderID = orderID
	return &order.Order{ID: orderID, Status: order.StatusConfirmed}, nil
}
func (f *fakePaymentService) ProcessManualPayment(ctx context.Context, req payment.ProcessManualPaymentRequest) (*payment.ProcessManualPaymentResponse, error) {
	f.lastProcessReq = req
	return &payment.ProcessManualPaymentResponse{}, nil
}
func (f *fakePaymentService) SimulatePay(ctx context.Context, orderID, method string) (*payment.Payment, error) {
	return nil, nil
}
func (f *fakePaymentService) SetSaleHook(hook payment.SaleHook) {}

// ---- tests ----

func TestOrderFacade_PassesTenantAndPagination(t *testing.T) {
	os := &fakeOrderService{}
	ps := &fakePaymentService{}
	f := NewOrderFacade(os, ps)

	_, err := f.ListOrders(context.Background(), "rst_42", nil, 25, 50)
	require.NoError(t, err)
	assert.Equal(t, "rst_42", os.lastListRestaurantID, "restaurant scope must be forwarded")
	assert.Equal(t, 25, os.lastListLimit)
	assert.Equal(t, 50, os.lastListOffset)
}

func TestOrderFacade_DeleteOrderForwardsRestaurantScope(t *testing.T) {
	os := &fakeOrderService{}
	f := NewOrderFacade(os, &fakePaymentService{})

	err := f.DeleteOrder(context.Background(), "ord_1", "rst_9")
	require.NoError(t, err)
	assert.Equal(t, "ord_1", os.lastDeleteOrderID)
	assert.Equal(t, "rst_9", os.lastDeleteRestoID, "delete must be scoped to the caller's restaurant")
}

func TestOrderFacade_ProcessPaymentUsesRequestIDs(t *testing.T) {
	ps := &fakePaymentService{}
	f := NewOrderFacade(&fakeOrderService{}, ps)

	_, err := f.ProcessPayment(context.Background(), payment.ProcessManualPaymentRequest{
		OrderID:      "ord_7",
		RestaurantID: "rst_jwt",
		CashierID:    "usr_cashier",
		PaidAmount:   20000,
	})
	require.NoError(t, err)
	assert.Equal(t, "ord_7", ps.lastProcessReq.OrderID)
	assert.Equal(t, "rst_jwt", ps.lastProcessReq.RestaurantID, "restaurant must come from the caller (JWT), not the body")
	assert.Equal(t, "usr_cashier", ps.lastProcessReq.CashierID)
}

func TestOrderFacade_CancelOrderSetsCancelledStatus(t *testing.T) {
	os := &fakeOrderService{}
	f := NewOrderFacade(os, &fakePaymentService{})

	err := f.CancelOrder(context.Background(), "ord_c", "usr_1")
	require.NoError(t, err)
	assert.Equal(t, "ord_c", os.lastCancelID)
	assert.Equal(t, order.StatusCancelled, os.lastCancelStatus)
}
