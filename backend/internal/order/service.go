package order

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"qr-store/backend/internal/menu"
	"qr-store/backend/internal/restaurant"
	"qr-store/backend/internal/table"
	"qr-store/backend/pkg/websocket"
)

var (
	ErrTableInactive       = errors.New("table is not active or not found")
	ErrRestaurantInactive  = errors.New("restaurant is not active")
	ErrMenuUnavailable     = errors.New("one or more items are unavailable")
	ErrEmptyOrder          = errors.New("order must contain at least one item")
	ErrOrderAccessDenied   = errors.New("order does not belong to this table session")
	ErrProofNotAllowed     = errors.New("proof of payment is not allowed for this order status")
	ErrInvalidProofURL     = errors.New("proof URL must point to an allowed image host")
)

// allowedProofHosts is the allowlist of hosts permitted for customer-uploaded
// payment proof images (Google Drive / Googleusercontent CDN).
var allowedProofHosts = map[string]bool{
	"lh3.googleusercontent.com": true,
	"drive.google.com":          true,
	"docs.google.com":           true,
	"script.googleusercontent.com": true,
}

type CreateOrderItemRequest struct {
	MenuID            string   `json:"menu_id" binding:"required"`
	Quantity          int      `json:"quantity" binding:"required,gt=0"`
	ModifierOptionIDs []string `json:"modifier_option_ids"`
	Notes             *string  `json:"notes"`
}

type CreateOrderRequest struct {
	QRToken string                   `json:"qr_token" binding:"required"`
	Items   []CreateOrderItemRequest `json:"items" binding:"required,min=1"`
	Notes   *string                  `json:"notes"`
}

type Service interface {
	CreatePublicOrder(ctx context.Context, req CreateOrderRequest) (*Order, error)
	GetOrderByID(ctx context.Context, id string) (*Order, error)
	GetPublicOrder(ctx context.Context, id, qrToken string) (*Order, error)
	ListOrders(ctx context.Context, restaurantID string, status *Status, limit, offset int) ([]Order, error)
	ListKitchenOrders(ctx context.Context, restaurantID string) ([]Order, error)
	UpdateOrderStatus(ctx context.Context, id string, newStatus Status, changedBy string) error
	SubmitProofOfPayment(ctx context.Context, orderID string, qrToken string, proofURL string) (*Order, error)
	DeleteOrder(ctx context.Context, id string, restaurantID string) error
	UpdateOrderDetails(ctx context.Context, id string, restaurantID string, notes *string, status *Status, paymentStatus *PaymentStatus) (*Order, error)
	GetTodayAnalytics(ctx context.Context, restaurantID string) (map[string]interface{}, error)
}

type service struct {
	orderRepo   Repository
	tableRepo   table.Repository
	restoRepo   restaurant.Repository
	menuRepo    menu.Repository
	hub         *websocket.Hub
}

func NewService(
	orderRepo Repository,
	tableRepo table.Repository,
	restoRepo restaurant.Repository,
	menuRepo menu.Repository,
	hub *websocket.Hub,
) Service {
	return &service{
		orderRepo: orderRepo,
		tableRepo: tableRepo,
		restoRepo: restoRepo,
		menuRepo:  menuRepo,
		hub:       hub,
	}
}

func (s *service) CreatePublicOrder(ctx context.Context, req CreateOrderRequest) (*Order, error) {
	if len(req.Items) == 0 {
		return nil, ErrEmptyOrder
	}

	// 1. Validate Table & Restaurant from QR Token
	tableInfo, err := s.tableRepo.GetByQRToken(ctx, req.QRToken)
	if err != nil {
		return nil, ErrTableInactive
	}

	restoID := tableInfo.Restaurant.ID
	tableID := tableInfo.Table.ID

	now := time.Now()
	var subtotal int64 = 0
	var orderItems []OrderItem

	// 2. Compute prices directly from database
	for _, itemReq := range req.Items {
		m, err := s.menuRepo.GetMenuByID(ctx, itemReq.MenuID, restoID)
		if err != nil {
			return nil, fmt.Errorf("menu item not found: %s", itemReq.MenuID)
		}
		if !m.Available {
			return nil, fmt.Errorf("%w: %s", ErrMenuUnavailable, m.Name)
		}

		unitPrice := m.Price
		var selectedModifiers []SelectedModifierOption

		// Build lookup map for modifier options of this menu
		optMap := make(map[string]struct {
			ModName  string
			OptName  string
			AddPrice int64
			ModID    string
		})
		for _, mod := range m.Modifiers {
			for _, opt := range mod.Options {
				optMap[opt.ID] = struct {
					ModName  string
					OptName  string
					AddPrice int64
					ModID    string
				}{
					ModName:  mod.Name,
					OptName:  opt.Name,
					AddPrice: opt.AdditionalPrice,
					ModID:    mod.ID,
				}
			}
		}

		for _, optID := range itemReq.ModifierOptionIDs {
			if opt, found := optMap[optID]; found {
				unitPrice += opt.AddPrice
				selectedModifiers = append(selectedModifiers, SelectedModifierOption{
					ModifierID:      opt.ModID,
					ModifierName:    opt.ModName,
					OptionID:        optID,
					OptionName:      opt.OptName,
					AdditionalPrice: opt.AddPrice,
				})
			}
		}

		itemSubtotal := unitPrice * int64(itemReq.Quantity)
		subtotal += itemSubtotal

		orderItems = append(orderItems, OrderItem{
			ID:                "item_" + uuid.New().String()[:8],
			MenuID:            &m.ID,
			MenuNameSnapshot:  m.Name,
			UnitPrice:         unitPrice,
			Quantity:          itemReq.Quantity,
			Subtotal:          itemSubtotal,
			SelectedModifiers: selectedModifiers,
			Notes:             itemReq.Notes,
			CreatedAt:         now,
		})
	}

	// 3. Tax and Service calculations
	tax, serviceCharge, total := computeTotals(subtotal, tableInfo.Restaurant.TaxPercent, tableInfo.Restaurant.ServicePercent)

	// 4. Generate human-readable order number with retry on collision.
	orderID := "ord_" + uuid.New().String()[:8]

	order := &Order{
		ID:            orderID,
		RestaurantID:  restoID,
		TableID:       tableID,
		TableName:     tableInfo.Table.Name,
		Status:        StatusWaitingPayment,
		PaymentStatus: PaymentStatusUnpaid,
		Subtotal:      subtotal,
		Tax:           tax,
		ServiceCharge: serviceCharge,
		Discount:      0,
		Total:         total,
		Notes:         req.Notes,
		Items:         orderItems,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	for i := range orderItems {
		orderItems[i].OrderID = order.ID
	}

	const maxOrderNumberAttempts = 5
	var createErr error
	for attempt := 0; attempt < maxOrderNumberAttempts; attempt++ {
		order.OrderNumber = generateOrderNumber(now)
		createErr = s.orderRepo.Create(ctx, order, orderItems)
		if createErr == nil {
			break
		}
		if !errors.Is(createErr, ErrOrderNumberConflict) {
			return nil, createErr
		}
	}
	if createErr != nil {
		return nil, createErr
	}

	if s.hub != nil {
		// Notify cashier room of incoming order
		s.hub.Publish(fmt.Sprintf("restaurant:%s:cashier", restoID), "NEW_ORDER_PENDING", order)
		s.hub.Publish(fmt.Sprintf("restaurant:%s:cashier", restoID), "order.created", order)
	}

	return order, nil
}

func (s *service) GetOrderByID(ctx context.Context, id string) (*Order, error) {
	return s.orderRepo.GetByID(ctx, id)
}

// generateOrderNumber returns a human-readable, collision-resistant order number
// of the form ORD-YYMMDD-XXXXXX using crypto/rand.
func generateOrderNumber(t time.Time) string {
	buf := make([]byte, 3)
	if _, err := rand.Read(buf); err != nil {
		// Fallback to nanoseconds if crypto/rand is unavailable (should not happen).
		return fmt.Sprintf("ORD-%s-%06d", t.Format("060102"), t.UnixNano()%1000000)
	}
	return fmt.Sprintf("ORD-%s-%s", t.Format("060102"), strings.ToUpper(hex.EncodeToString(buf)))
}

// computeTotals derives tax, service charge and grand total from a subtotal
// using the restaurant's percentage configuration. Kept as a pure function so
// the money math can be unit-tested without database dependencies.
func computeTotals(subtotal int64, taxPercent, servicePercent float64) (tax, serviceCharge, total int64) {
	tax = int64(float64(subtotal) * (taxPercent / 100.0))
	serviceCharge = int64(float64(subtotal) * (servicePercent / 100.0))
	total = subtotal + tax + serviceCharge
	return tax, serviceCharge, total
}

// GetPublicOrder returns an order only when the supplied QR token maps to the
// table the order belongs to. This prevents IDOR on the public tracking route.
func (s *service) GetPublicOrder(ctx context.Context, id, qrToken string) (*Order, error) {
	if qrToken == "" {
		return nil, ErrOrderAccessDenied
	}
	o, err := s.orderRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	tableInfo, err := s.tableRepo.GetByQRToken(ctx, qrToken)
	if err != nil {
		return nil, ErrOrderAccessDenied
	}
	if tableInfo.Table.ID != o.TableID || tableInfo.Restaurant.ID != o.RestaurantID {
		return nil, ErrOrderAccessDenied
	}
	return o, nil
}

func (s *service) ListOrders(ctx context.Context, restaurantID string, status *Status, limit, offset int) ([]Order, error) {
	return s.orderRepo.List(ctx, restaurantID, status, limit, offset)
}

func (s *service) ListKitchenOrders(ctx context.Context, restaurantID string) ([]Order, error) {
	return s.orderRepo.ListKitchenOrders(ctx, restaurantID)
}

func (s *service) UpdateOrderStatus(ctx context.Context, id string, newStatus Status, changedBy string) error {
	if err := s.orderRepo.UpdateStatus(ctx, id, newStatus, &changedBy); err != nil {
		return err
	}

	o, err := s.orderRepo.GetByID(ctx, id)
	if err == nil && s.hub != nil {
		// Broadcast to customer tracking room
		s.hub.Publish(fmt.Sprintf("order:%s", id), "ORDER_STATUS_CHANGED", map[string]interface{}{
			"order_id": o.ID,
			"status":   o.Status,
			"order":    o,
		})

		// Broadcast to kitchen room
		s.hub.Publish(fmt.Sprintf("restaurant:%s:kitchen", o.RestaurantID), "ORDER_STATUS_CHANGED", map[string]interface{}{
			"order_id": o.ID,
			"status":   o.Status,
			"order":    o,
		})

		// Broadcast to cashier room
		s.hub.Publish(fmt.Sprintf("restaurant:%s:cashier", o.RestaurantID), "ORDER_STATUS_CHANGED", map[string]interface{}{
			"order_id": o.ID,
			"status":   o.Status,
			"order":    o,
		})
	}

	return nil
}

func (s *service) SubmitProofOfPayment(ctx context.Context, orderID string, qrToken string, proofURL string) (*Order, error) {
	if !isAllowedProofURL(proofURL) {
		return nil, ErrInvalidProofURL
	}

	o, err := s.GetPublicOrder(ctx, orderID, qrToken)
	if err != nil {
		return nil, err
	}

	// Only orders awaiting payment may receive proof.
	if o.Status != StatusWaitingPayment && o.Status != StatusPendingConfirmation {
		return nil, ErrProofNotAllowed
	}

	if err := s.orderRepo.UpdateProofURL(ctx, orderID, proofURL); err != nil {
		return nil, err
	}

	o, err = s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return nil, err
	}

	if s.hub != nil {
		// Broadcast to cashier and customer rooms
		eventData := map[string]interface{}{
			"order_id":  o.ID,
			"order":     o,
			"proof_url": proofURL,
		}
		s.hub.Publish(fmt.Sprintf("restaurant:%s:cashier", o.RestaurantID), "PAYMENT_PROOF_SUBMITTED", eventData)
		s.hub.Publish(fmt.Sprintf("order:%s", o.ID), "PAYMENT_PROOF_SUBMITTED", eventData)
	}

	return o, nil
}

// isAllowedProofURL validates that proofURL is an http(s) URL on an allowed host.
func isAllowedProofURL(proofURL string) bool {
	u, err := url.Parse(proofURL)
	if err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" {
		return false
	}
	host := strings.ToLower(u.Hostname())
	return allowedProofHosts[host]
}

func (s *service) GetTodayAnalytics(ctx context.Context, restaurantID string) (map[string]interface{}, error) {
	rev, count, avg, paid, cancelled, err := s.orderRepo.GetTodayStats(ctx, restaurantID)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"today_revenue":    rev,
		"today_orders":     count,
		"today_paid":       paid,
		"today_cancelled":  cancelled,
		"average_order":    avg,
	}, nil
}

func (s *service) DeleteOrder(ctx context.Context, id string, restaurantID string) error {
	o, _ := s.orderRepo.GetByID(ctx, id)
	if err := s.orderRepo.Delete(ctx, id, restaurantID); err != nil {
		return err
	}

	if s.hub != nil {
		eventData := map[string]interface{}{
			"order_id":      id,
			"restaurant_id": restaurantID,
			"deleted":       true,
		}
		if o != nil {
			eventData["order_number"] = o.OrderNumber
			eventData["table_id"] = o.TableID
		}
		s.hub.Publish(fmt.Sprintf("restaurant:%s:cashier", restaurantID), "ORDER_DELETED", eventData)
		s.hub.Publish(fmt.Sprintf("restaurant:%s:kitchen", restaurantID), "ORDER_DELETED", eventData)
		s.hub.Publish(fmt.Sprintf("order:%s", id), "ORDER_DELETED", eventData)
	}
	return nil
}

func (s *service) UpdateOrderDetails(ctx context.Context, id string, restaurantID string, notes *string, status *Status, paymentStatus *PaymentStatus) (*Order, error) {
	if err := s.orderRepo.UpdateDetails(ctx, id, notes, status, paymentStatus); err != nil {
		return nil, err
	}

	o, err := s.orderRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if s.hub != nil {
		eventData := map[string]interface{}{
			"order_id": o.ID,
			"status":   o.Status,
			"order":    o,
		}
		s.hub.Publish(fmt.Sprintf("restaurant:%s:cashier", restaurantID), "ORDER_STATUS_CHANGED", eventData)
		s.hub.Publish(fmt.Sprintf("restaurant:%s:kitchen", restaurantID), "ORDER_STATUS_CHANGED", eventData)
		s.hub.Publish(fmt.Sprintf("order:%s", id), "ORDER_STATUS_CHANGED", eventData)
	}
	return o, nil
}
