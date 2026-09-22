package payment

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"qr-store/backend/internal/order"
	"qr-store/backend/pkg/logger"
	"qr-store/backend/pkg/websocket"
)

var (
	ErrOrderAlreadyPaid = errors.New("order is already paid")
	ErrInvalidAmount    = errors.New("webhook amount does not match order total")
	ErrWebhookDuplicate = errors.New("webhook event already processed")
)

type SaleHook interface {
	RecordSaleForOrder(ctx context.Context, restaurantID, orderID, paymentID string, amount int64, method string) error
}

type Service interface {
	CreatePaymentForOrder(ctx context.Context, orderID string) (*Payment, error)
	GetPaymentByID(ctx context.Context, id string) (*Payment, error)
	GetPaymentByOrderID(ctx context.Context, orderID string) (*Payment, error)
	ProcessWebhook(ctx context.Context, headers http.Header, rawBody []byte) (*WebhookResult, error)
	SimulatePay(ctx context.Context, orderID string) (*Payment, error)
	SetSaleHook(hook SaleHook)
}

type service struct {
	paymentRepo Repository
	orderRepo   order.Repository
	gateway     PaymentGateway
	hub         *websocket.Hub
	saleHook    SaleHook
}

func NewService(
	paymentRepo Repository,
	orderRepo order.Repository,
	gateway PaymentGateway,
	hub *websocket.Hub,
) Service {
	return &service{
		paymentRepo: paymentRepo,
		orderRepo:   orderRepo,
		gateway:     gateway,
		hub:         hub,
	}
}

func (s *service) SetSaleHook(hook SaleHook) {
	s.saleHook = hook
}

func (s *service) CreatePaymentForOrder(ctx context.Context, orderID string) (*Payment, error) {
	o, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return nil, err
	}

	if o.Status != order.StatusWaitingPayment {
		return nil, ErrOrderAlreadyPaid
	}

	// Check if there is already an active pending payment
	existing, err := s.paymentRepo.GetByOrderID(ctx, orderID)
	if err == nil && existing != nil && existing.Status == StatusPending {
		if existing.ExpiredAt == nil || existing.ExpiredAt.After(time.Now()) {
			return existing, nil
		}
	}

	// Request new payment from gateway
	gwResp, err := s.gateway.CreatePayment(ctx, CreatePaymentRequest{
		OrderID:     o.ID,
		OrderNumber: o.OrderNumber,
		Amount:      o.Total,
		ItemName:    fmt.Sprintf("Order %s", o.OrderNumber),
	})
	if err != nil {
		return nil, fmt.Errorf("gateway error: %w", err)
	}

	now := time.Now()
	p := &Payment{
		ID:                    "pay_" + uuid.New().String()[:8],
		RestaurantID:          o.RestaurantID,
		OrderID:               o.ID,
		Provider:              s.gateway.Name(),
		ProviderTransactionID: &gwResp.TransactionID,
		PaymentMethod:         nil,
		Amount:                o.Total,
		Status:                StatusPending,
		PaymentURL:            &gwResp.PaymentURL,
		QRString:              &gwResp.QRString,
		ExpiredAt:             &gwResp.ExpiredAt,
		CreatedAt:             now,
		UpdatedAt:             now,
	}

	if err := s.paymentRepo.Create(ctx, p); err != nil {
		return nil, err
	}

	return p, nil
}

func (s *service) GetPaymentByID(ctx context.Context, id string) (*Payment, error) {
	return s.paymentRepo.GetByID(ctx, id)
}

func (s *service) GetPaymentByOrderID(ctx context.Context, orderID string) (*Payment, error) {
	return s.paymentRepo.GetByOrderID(ctx, orderID)
}

func (s *service) ProcessWebhook(ctx context.Context, headers http.Header, rawBody []byte) (*WebhookResult, error) {
	result, err := s.gateway.VerifyWebhook(headers, rawBody)
	if err != nil {
		return nil, fmt.Errorf("signature verification failed: %w", err)
	}

	// Enforce Webhook Idempotency (PRD Section 39, Section 71 Case 2)
	now := time.Now()
	event := &WebhookEvent{
		ID:        "whe_" + uuid.New().String()[:8],
		Provider:  s.gateway.Name(),
		EventID:   result.EventID,
		EventType: &result.EventType,
		Payload:   string(rawBody),
		Processed: true,
		CreatedAt: now,
	}

	isDuplicate, err := s.paymentRepo.RecordWebhookEvent(ctx, event)
	if err != nil {
		return nil, fmt.Errorf("failed to record webhook event: %w", err)
	}
	if isDuplicate {
		logger.Log.Info("duplicate webhook event ignored", "event_id", result.EventID, "provider", s.gateway.Name())
		return result, nil // Idempotent success response
	}

	// Fetch Order & Validate Amount (Rule 4 & Case 3)
	o, err := s.orderRepo.GetByID(ctx, result.OrderID)
	if err != nil {
		return nil, fmt.Errorf("order not found for webhook: %w", err)
	}

	if result.Amount != o.Total {
		logger.Log.Error("amount mismatch", "expected", o.Total, "received", result.Amount)
		return nil, ErrInvalidAmount
	}

	// Update Payment and Order state
	if result.Status == StatusPaid {
		paymentRecord, err := s.paymentRepo.GetByOrderID(ctx, o.ID)
		paymentID := ""
		if err == nil && paymentRecord != nil {
			paymentID = paymentRecord.ID
		}

		paidAt := now
		if result.PaidAt != nil {
			paidAt = *result.PaidAt
		}

		updatedOrder, err := s.paymentRepo.ProcessPaymentSuccess(
			ctx, o.ID, paymentID, paidAt, result.PaymentMethod, result.ProviderTransactionID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to settle payment in database: %w", err)
		}

		// Hook into active cashier shift if exists
		if s.saleHook != nil {
			_ = s.saleHook.RecordSaleForOrder(ctx, o.RestaurantID, o.ID, paymentID, o.Total, result.PaymentMethod)
		}

		// Publish real-time events to customer and kitchen (PRD Section 38, 49)
		if s.hub != nil {
			s.hub.Publish(fmt.Sprintf("order:%s", o.ID), "PAYMENT_PAID", map[string]interface{}{
				"order_id": o.ID,
				"status":   order.StatusConfirmed,
				"order":    updatedOrder,
			})

			s.hub.Publish(fmt.Sprintf("restaurant:%s:kitchen", o.RestaurantID), "NEW_ORDER_CONFIRMED", map[string]interface{}{
				"order_id": o.ID,
				"status":   order.StatusConfirmed,
				"order":    updatedOrder,
			})
		}
	}

	return result, nil
}

func (s *service) SimulatePay(ctx context.Context, orderID string) (*Payment, error) {
	o, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return nil, err
	}

	if o.Status != order.StatusWaitingPayment {
		return nil, ErrOrderAlreadyPaid
	}

	// Create payment if not exists
	p, err := s.CreatePaymentForOrder(ctx, orderID)
	if err != nil {
		return nil, err
	}

	// Build simulated mock webhook payload
	mockPayload := MockWebhookPayload{
		EventID:       "sim_" + uuid.New().String()[:8],
		OrderID:       o.ID,
		TransactionID: "tx_" + uuid.New().String()[:8],
		Amount:        o.Total,
		Status:        "PAID",
		PaymentMethod: "QRIS",
		Signature:     "valid",
	}
	body, _ := json.Marshal(mockPayload)

	headers := make(http.Header)
	headers.Set("Content-Type", "application/json")

	_, err = s.ProcessWebhook(ctx, headers, body)
	if err != nil {
		return nil, err
	}

	return s.paymentRepo.GetByID(ctx, p.ID)
}
