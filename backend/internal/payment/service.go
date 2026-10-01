package payment

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"qr-store/backend/internal/order"
	"qr-store/backend/pkg/logger"
	"qr-store/backend/pkg/websocket"
)

var (
	ErrOrderAlreadyPaid       = errors.New("order is already paid")
	ErrInvalidAmount          = errors.New("webhook amount does not match order total")
	ErrWebhookDuplicate       = errors.New("webhook event already processed")
	ErrInsufficientPaidAmount = errors.New("paid amount is less than order total")
	ErrInvalidPaymentMethod   = errors.New("invalid payment method")
	ErrOrderCancelled         = errors.New("cannot process payment for cancelled order")
)

type ProcessManualPaymentRequest struct {
	OrderID         string  `json:"order_id" binding:"required"`
	RestaurantID    string  `json:"restaurant_id"`
	CashierID       string  `json:"cashier_id"`
	PaymentMethod   string  `json:"payment_method" binding:"required"`
	PaidAmount      int64   `json:"paid_amount"`
	ReferenceNumber *string `json:"reference_number"`
}

type ProcessManualPaymentResponse struct {
	Payment *Payment     `json:"payment"`
	Order   *order.Order `json:"order"`
	Change  int64        `json:"change"`
}

type SaleHook interface {
	RecordSaleForOrder(ctx context.Context, restaurantID, orderID, paymentID string, amount int64, method string) error
}

type Service interface {
	CreatePaymentForOrder(ctx context.Context, orderID string) (*Payment, error)
	GetPaymentByID(ctx context.Context, id string) (*Payment, error)
	GetPaymentByOrderID(ctx context.Context, orderID string) (*Payment, error)
	ProcessWebhook(ctx context.Context, headers http.Header, rawBody []byte) (*WebhookResult, error)
	SimulatePay(ctx context.Context, orderID string, paymentMethod string) (*Payment, error)
	ProcessManualPayment(ctx context.Context, req ProcessManualPaymentRequest) (*ProcessManualPaymentResponse, error)
	ConfirmOrder(ctx context.Context, orderID string, confirmedBy string) (*order.Order, error)
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

func (s *service) SimulatePay(ctx context.Context, orderID string, paymentMethod string) (*Payment, error) {
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

	method := paymentMethod
	if method == "" {
		method = "QRIS"
	}

	// Build simulated mock webhook payload
	mockPayload := MockWebhookPayload{
		EventID:       "sim_" + uuid.New().String()[:8],
		OrderID:       o.ID,
		TransactionID: "tx_" + uuid.New().String()[:8],
		Amount:        o.Total,
		Status:        "PAID",
		PaymentMethod: method,
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

func (s *service) ProcessManualPayment(ctx context.Context, req ProcessManualPaymentRequest) (*ProcessManualPaymentResponse, error) {
	o, err := s.orderRepo.GetByID(ctx, req.OrderID)
	if err != nil {
		return nil, err
	}

	if req.RestaurantID != "" && o.RestaurantID != req.RestaurantID {
		return nil, errors.New("order does not belong to this restaurant")
	}

	if o.Status == order.StatusCancelled {
		return nil, ErrOrderCancelled
	}

	// Idempotency: if already paid, return existing payment
	if o.PaymentStatus == order.PaymentStatusPaid {
		existing, err := s.paymentRepo.GetByOrderID(ctx, req.OrderID)
		if err == nil && existing != nil {
			return &ProcessManualPaymentResponse{
				Payment: existing,
				Order:   o,
				Change:  existing.ChangeAmount,
			}, nil
		}
		return nil, ErrOrderAlreadyPaid
	}

	// Validate & normalize payment method
	method := strings.ToUpper(strings.TrimSpace(req.PaymentMethod))
	if method != order.PaymentMethodCash && method != order.PaymentMethodQRISManual && method != order.PaymentMethodDebit && method != order.PaymentMethodOther && method != "QRIS" {
		return nil, fmt.Errorf("%w: %s", ErrInvalidPaymentMethod, req.PaymentMethod)
	}
	if method == "QRIS" {
		method = order.PaymentMethodQRISManual
	}

	// Cash calculation & validation (PRD Section 9: change = paid_amount - grand_total)
	var paidAmount int64
	var change int64 = 0

	if method == order.PaymentMethodCash {
		if req.PaidAmount < o.Total {
			return nil, fmt.Errorf("%w: paid %d < total %d", ErrInsufficientPaidAmount, req.PaidAmount, o.Total)
		}
		paidAmount = req.PaidAmount
		change = paidAmount - o.Total
	} else {
		// Non-cash (QRIS_MANUAL, DEBIT, OTHER)
		if req.PaidAmount <= 0 {
			paidAmount = o.Total
		} else {
			paidAmount = req.PaidAmount
		}
		change = 0
	}

	now := time.Now()

	// Check existing payment record
	existing, err := s.paymentRepo.GetByOrderID(ctx, o.ID)
	var p *Payment
	if err == nil && existing != nil {
		p = existing
		p.Provider = "MANUAL"
		p.PaymentMethod = &method
		p.PaidAmount = paidAmount
		p.ChangeAmount = change
		p.ReferenceNumber = req.ReferenceNumber
		p.Status = StatusPaid
		p.VerifiedBy = &req.CashierID
		p.VerifiedAt = &now
		p.PaidAt = &now
		p.UpdatedAt = now
		if err := s.paymentRepo.Update(ctx, p); err != nil {
			return nil, fmt.Errorf("failed to update payment record: %w", err)
		}
	} else {
		p = &Payment{
			ID:              "pay_" + uuid.New().String()[:8],
			RestaurantID:    o.RestaurantID,
			OrderID:         o.ID,
			Provider:        "MANUAL",
			PaymentMethod:   &method,
			Amount:          o.Total,
			PaidAmount:      paidAmount,
			ChangeAmount:    change,
			ReferenceNumber: req.ReferenceNumber,
			Status:          StatusPaid,
			VerifiedBy:      &req.CashierID,
			VerifiedAt:      &now,
			PaidAt:          &now,
			CreatedAt:       now,
			UpdatedAt:       now,
		}
		if err := s.paymentRepo.Create(ctx, p); err != nil {
			return nil, fmt.Errorf("failed to create payment record: %w", err)
		}
	}

	// Settle order status in database
	updatedOrder, err := s.paymentRepo.ProcessPaymentSuccess(ctx, o.ID, p.ID, now, method, p.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to settle order: %w", err)
	}

	// Record sale in shift ledger if hook is configured
	if s.saleHook != nil {
		_ = s.saleHook.RecordSaleForOrder(ctx, o.RestaurantID, o.ID, p.ID, o.Total, method)
	}

	// Publish realtime WebSocket events
	if s.hub != nil {
		eventData := map[string]interface{}{
			"order_id":       o.ID,
			"status":         updatedOrder.Status,
			"payment_status": updatedOrder.PaymentStatus,
			"payment_method": method,
			"paid_amount":    paidAmount,
			"change_amount":  change,
			"order":          updatedOrder,
			"payment":        p,
		}

		// Notify customer tracking room
		s.hub.Publish(fmt.Sprintf("order:%s", o.ID), "PAYMENT_PAID", eventData)
		s.hub.Publish(fmt.Sprintf("order:%s", o.ID), "order.confirmed", eventData)

		// Notify kitchen room
		s.hub.Publish(fmt.Sprintf("restaurant:%s:kitchen", o.RestaurantID), "NEW_ORDER_CONFIRMED", eventData)
		s.hub.Publish(fmt.Sprintf("restaurant:%s:kitchen", o.RestaurantID), "kitchen.new_order", eventData)

		// Notify cashier room
		s.hub.Publish(fmt.Sprintf("restaurant:%s:cashier", o.RestaurantID), "PAYMENT_CONFIRMED", eventData)
		s.hub.Publish(fmt.Sprintf("restaurant:%s:cashier", o.RestaurantID), "order.confirmed", eventData)
	}

	return &ProcessManualPaymentResponse{
		Payment: p,
		Order:   updatedOrder,
		Change:  change,
	}, nil
}

func (s *service) ConfirmOrder(ctx context.Context, orderID string, confirmedBy string) (*order.Order, error) {
	o, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return nil, err
	}

	if o.Status == order.StatusConfirmed {
		return o, nil
	}

	if err := s.orderRepo.UpdateStatus(ctx, orderID, order.StatusConfirmed, &confirmedBy); err != nil {
		return nil, err
	}

	updated, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return nil, err
	}

	if s.hub != nil {
		s.hub.Publish(fmt.Sprintf("order:%s", orderID), "ORDER_CONFIRMED", updated)
		s.hub.Publish(fmt.Sprintf("restaurant:%s:kitchen", updated.RestaurantID), "NEW_ORDER_CONFIRMED", updated)
		s.hub.Publish(fmt.Sprintf("restaurant:%s:cashier", updated.RestaurantID), "ORDER_CONFIRMED", updated)
	}

	return updated, nil
}
