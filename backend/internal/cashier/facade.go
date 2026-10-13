package cashier

import (
	"context"

	"qr-store/backend/internal/order"
	"qr-store/backend/internal/payment"
)

// OrderFacade is an application-facing boundary that composes the order and
// payment domains for cashier workflows. It exists so the HTTP handler does not
// orchestrate multiple domain services directly (layer separation, see plan
// item 2.8): the handler speaks only to this facade.
type OrderFacade struct {
	orderService   order.Service
	paymentService payment.Service
}

// NewOrderFacade builds the facade from the order & payment services.
func NewOrderFacade(orderService order.Service, paymentService payment.Service) *OrderFacade {
	return &OrderFacade{
		orderService:   orderService,
		paymentService: paymentService,
	}
}

// ListOrders returns a page of orders for the restaurant.
func (f *OrderFacade) ListOrders(ctx context.Context, restaurantID string, status *order.Status, limit, offset int) ([]order.Order, error) {
	return f.orderService.ListOrders(ctx, restaurantID, status, limit, offset)
}

// GetOrder fetches a single order by ID.
func (f *OrderFacade) GetOrder(ctx context.Context, id string) (*order.Order, error) {
	return f.orderService.GetOrderByID(ctx, id)
}

// ConfirmOrder confirms an order (WAITING_PAYMENT -> CONFIRMED).
func (f *OrderFacade) ConfirmOrder(ctx context.Context, orderID, confirmedBy string) (*order.Order, error) {
	return f.paymentService.ConfirmOrder(ctx, orderID, confirmedBy)
}

// ProcessPayment settles a cashier manual payment for an order.
func (f *OrderFacade) ProcessPayment(ctx context.Context, req payment.ProcessManualPaymentRequest) (*payment.ProcessManualPaymentResponse, error) {
	return f.paymentService.ProcessManualPayment(ctx, req)
}

// CancelOrder transitions an order to CANCELLED.
func (f *OrderFacade) CancelOrder(ctx context.Context, orderID, changedBy string) error {
	return f.orderService.UpdateOrderStatus(ctx, orderID, order.StatusCancelled, changedBy)
}

// DeleteOrder removes an order scoped to the restaurant.
func (f *OrderFacade) DeleteOrder(ctx context.Context, orderID, restaurantID string) error {
	return f.orderService.DeleteOrder(ctx, orderID, restaurantID)
}
