package cashier

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"qr-store/backend/internal/order"
	"qr-store/backend/internal/payment"
	"qr-store/backend/pkg/logger"
	"qr-store/backend/pkg/websocket"
)

var (
	ErrOpeningBalanceNegative = errors.New("opening balance cannot be negative")
	ErrRegisterInactive       = errors.New("register is not active")
	ErrOrderNotEligibleRefund = errors.New("only paid orders can be refunded")
	ErrOrderNotEligibleVoid   = errors.New("completed orders cannot be voided")
	ErrRefundAmountExceeds    = errors.New("refund amount exceeds order total")
)

type CreateRegisterRequest struct {
	Name string `json:"name" binding:"required"`
}

type UpdateRegisterRequest struct {
	Name   *string `json:"name"`
	Status *string `json:"status"`
}

type OpenShiftRequest struct {
	RegisterID     string `json:"register_id" binding:"required"`
	OpeningBalance int64  `json:"opening_balance"`
}

type CloseShiftRequest struct {
	ActualAmount *int64  `json:"actual_amount"`
	Notes        *string `json:"notes"`
}

type CreateRefundRequest struct {
	Amount int64  `json:"amount" binding:"required"`
	Reason string `json:"reason" binding:"required"`
}

type CreateVoidRequest struct {
	Reason string `json:"reason" binding:"required"`
}

type ShiftReportDetail struct {
	Shift        CashierShift              `json:"shift"`
	Summary      ShiftSummary              `json:"summary"`
	Closing      *ShiftClosing             `json:"closing,omitempty"`
	Transactions []CashierShiftTransaction `json:"transactions"`
}

type Service interface {
	// Register
	ListRegisters(ctx context.Context, restaurantID string) ([]Register, error)
	GetRegister(ctx context.Context, id string) (*Register, error)
	CreateRegister(ctx context.Context, restaurantID string, req CreateRegisterRequest) (*Register, error)
	UpdateRegister(ctx context.Context, id string, req UpdateRegisterRequest) (*Register, error)
	DeleteRegister(ctx context.Context, id string) error

	// Shift
	OpenShift(ctx context.Context, restaurantID, cashierID string, req OpenShiftRequest) (*CashierShift, error)
	GetCurrentShift(ctx context.Context, restaurantID, cashierID string) (*CashierShift, error)
	GetShiftByID(ctx context.Context, id string) (*CashierShift, error)
	GetShiftTransactions(ctx context.Context, shiftID string, txType *TransactionType, page, limit int) ([]CashierShiftTransaction, int, error)
	GetShiftSummary(ctx context.Context, shiftID string) (*ShiftSummary, error)
	CloseShift(ctx context.Context, shiftID, cashierID string, req CloseShiftRequest) (*ShiftClosing, error)

	// Payment / Order hooks
	RecordSaleForOrder(ctx context.Context, restaurantID, orderID, paymentID string, amount int64, method string) error

	// Refund / Void
	CreateRefund(ctx context.Context, restaurantID, orderID, userID string, req CreateRefundRequest) (*Refund, error)
	CreateVoid(ctx context.Context, restaurantID, orderID, userID string, req CreateVoidRequest) error

	// Reports
	ListReports(ctx context.Context, restaurantID string, filters ShiftFilter) ([]CashierShift, error)
	GetReportDetail(ctx context.Context, shiftID string) (*ShiftReportDetail, error)
}

type service struct {
	cashierRepo Repository
	orderRepo   order.Repository
	paymentRepo payment.Repository
	hub         *websocket.Hub
}

func NewService(
	cashierRepo Repository,
	orderRepo order.Repository,
	paymentRepo payment.Repository,
	hub *websocket.Hub,
) Service {
	return &service{
		cashierRepo: cashierRepo,
		orderRepo:   orderRepo,
		paymentRepo: paymentRepo,
		hub:         hub,
	}
}

// ----------------- REGISTERS -----------------

func (s *service) ListRegisters(ctx context.Context, restaurantID string) ([]Register, error) {
	return s.cashierRepo.ListRegisters(ctx, restaurantID)
}

func (s *service) GetRegister(ctx context.Context, id string) (*Register, error) {
	return s.cashierRepo.GetRegisterByID(ctx, id)
}

func (s *service) CreateRegister(ctx context.Context, restaurantID string, req CreateRegisterRequest) (*Register, error) {
	now := time.Now()
	reg := &Register{
		ID:           "reg_" + uuid.New().String()[:8],
		RestaurantID: restaurantID,
		Name:         req.Name,
		Status:       "ACTIVE",
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.cashierRepo.CreateRegister(ctx, reg); err != nil {
		return nil, err
	}
	return reg, nil
}

func (s *service) UpdateRegister(ctx context.Context, id string, req UpdateRegisterRequest) (*Register, error) {
	reg, err := s.cashierRepo.GetRegisterByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if req.Name != nil {
		reg.Name = *req.Name
	}
	if req.Status != nil {
		reg.Status = *req.Status
	}
	if err := s.cashierRepo.UpdateRegister(ctx, reg); err != nil {
		return nil, err
	}
	return reg, nil
}

func (s *service) DeleteRegister(ctx context.Context, id string) error {
	return s.cashierRepo.DeleteRegister(ctx, id)
}

// ----------------- SHIFTS -----------------

func (s *service) OpenShift(ctx context.Context, restaurantID, cashierID string, req OpenShiftRequest) (*CashierShift, error) {
	if req.OpeningBalance < 0 {
		return nil, ErrOpeningBalanceNegative
	}

	reg, err := s.cashierRepo.GetRegisterByID(ctx, req.RegisterID)
	if err != nil {
		return nil, err
	}
	if reg.Status != "ACTIVE" {
		return nil, ErrRegisterInactive
	}

	now := time.Now()
	shift := &CashierShift{
		ID:             "shf_" + uuid.New().String()[:8],
		RestaurantID:   restaurantID,
		RegisterID:     req.RegisterID,
		CashierID:      cashierID,
		Status:         ShiftStatusOpen,
		OpeningBalance: req.OpeningBalance,
		OpenedAt:       now,
		ExpectedTotal:  req.OpeningBalance,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := s.cashierRepo.CreateShift(ctx, shift); err != nil {
		return nil, err
	}

	// Audit Log
	_ = s.cashierRepo.CreateAuditLog(ctx, restaurantID, cashierID, "CASHIER_SHIFT_OPENED", "cashier_shifts", shift.ID, map[string]interface{}{
		"register_id":     shift.RegisterID,
		"opening_balance": shift.OpeningBalance,
	})

	if s.hub != nil {
		s.hub.Publish(fmt.Sprintf("restaurant:%s:cashier", restaurantID), "CASHIER_SHIFT_OPENED", shift)
	}

	return s.cashierRepo.GetShiftByID(ctx, shift.ID)
}

func (s *service) GetCurrentShift(ctx context.Context, restaurantID, cashierID string) (*CashierShift, error) {
	// First look for this cashier's active shift
	shift, err := s.cashierRepo.GetActiveShiftByCashier(ctx, cashierID)
	if err != nil {
		return nil, err
	}
	if shift != nil {
		return shift, nil
	}

	// Fallback to active shift in restaurant
	return s.cashierRepo.GetActiveShiftByRestaurant(ctx, restaurantID)
}

func (s *service) GetShiftByID(ctx context.Context, id string) (*CashierShift, error) {
	return s.cashierRepo.GetShiftByID(ctx, id)
}

func (s *service) GetShiftTransactions(ctx context.Context, shiftID string, txType *TransactionType, page, limit int) ([]CashierShiftTransaction, int, error) {
	return s.cashierRepo.ListTransactions(ctx, shiftID, txType, page, limit)
}

func (s *service) GetShiftSummary(ctx context.Context, shiftID string) (*ShiftSummary, error) {
	return s.cashierRepo.GetShiftSummary(ctx, shiftID)
}

func (s *service) CloseShift(ctx context.Context, shiftID, cashierID string, req CloseShiftRequest) (*ShiftClosing, error) {
	shift, err := s.cashierRepo.GetShiftByID(ctx, shiftID)
	if err != nil {
		return nil, err
	}
	if shift.Status == ShiftStatusClosed {
		return nil, ErrShiftAlreadyClosed
	}

	// Calculate final summary
	summary, err := s.cashierRepo.GetShiftSummary(ctx, shiftID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	var difference *int64
	if req.ActualAmount != nil {
		diff := CalculateDifference(summary.ExpectedAmount, *req.ActualAmount)
		difference = &diff
	}

	shift.Status = ShiftStatusClosed
	shift.ClosedAt = &now
	shift.ExpectedTotal = summary.ExpectedAmount
	shift.ActualTotal = req.ActualAmount
	shift.Difference = difference

	closing := &ShiftClosing{
		ID:              "cls_" + uuid.New().String()[:8],
		ShiftID:         shift.ID,
		GrossSales:      summary.GrossSales,
		RefundTotal:     summary.RefundTotal,
		VoidTotal:       summary.VoidTotal,
		AdjustmentTotal: summary.AdjustmentTotal,
		NetSales:        summary.NetSales,
		ExpectedAmount:  summary.ExpectedAmount,
		ActualAmount:    req.ActualAmount,
		Difference:      difference,
		Notes:           req.Notes,
		ClosedBy:        cashierID,
		ClosedAt:        now,
	}

	if err := s.cashierRepo.CloseShiftTx(ctx, shift, closing); err != nil {
		return nil, err
	}

	// Audit Log
	_ = s.cashierRepo.CreateAuditLog(ctx, shift.RestaurantID, cashierID, "CASHIER_SHIFT_CLOSED", "cashier_shifts", shift.ID, map[string]interface{}{
		"net_sales":       summary.NetSales,
		"expected_amount": summary.ExpectedAmount,
		"actual_amount":   req.ActualAmount,
		"difference":      difference,
	})

	if s.hub != nil {
		s.hub.Publish(fmt.Sprintf("restaurant:%s:cashier", shift.RestaurantID), "CASHIER_SHIFT_CLOSED", closing)
	}

	return closing, nil
}

// ----------------- SALE HOOK -----------------

func (s *service) RecordSaleForOrder(ctx context.Context, restaurantID, orderID, paymentID string, amount int64, method string) error {
	shift, err := s.cashierRepo.GetActiveShiftByRestaurant(ctx, restaurantID)
	if err != nil || shift == nil {
		// Self-order without active shift is allowed per PRD!
		logger.Log.Info("no active shift for restaurant, skipping shift transaction record", "restaurant_id", restaurantID)
		return nil
	}

	now := time.Now()
	tx := &CashierShiftTransaction{
		ID:        "cst_" + uuid.New().String()[:8],
		ShiftID:   shift.ID,
		OrderID:   &orderID,
		PaymentID: &paymentID,
		Type:      TxTypeSale,
		Amount:    amount,
		Metadata: map[string]interface{}{
			"payment_method": method,
		},
		CreatedAt: now,
	}

	if err := s.cashierRepo.CreateTransaction(ctx, tx); err != nil {
		logger.Log.Error("failed to record sale in shift transaction", "error", err)
		return err
	}

	if s.hub != nil {
		s.hub.Publish(fmt.Sprintf("shift:%s", shift.ID), "SHIFT_TRANSACTION_ADDED", tx)
	}

	return nil
}

// ----------------- REFUND / VOID -----------------

func (s *service) CreateRefund(ctx context.Context, restaurantID, orderID, userID string, req CreateRefundRequest) (*Refund, error) {
	o, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if o.Status != order.StatusConfirmed && o.Status != order.StatusPreparing && o.Status != order.StatusReady && o.Status != order.StatusCompleted {
		return nil, ErrOrderNotEligibleRefund
	}

	if req.Amount > o.Total {
		return nil, ErrRefundAmountExceeds
	}

	p, err := s.paymentRepo.GetByOrderID(ctx, orderID)
	if err != nil || p == nil {
		return nil, errors.New("no payment found for this order")
	}

	now := time.Now()
	ref := &Refund{
		ID:           "ref_" + uuid.New().String()[:8],
		RestaurantID: restaurantID,
		OrderID:      orderID,
		PaymentID:    p.ID,
		Amount:       req.Amount,
		Reason:       req.Reason,
		Status:       RefundStatusCompleted,
		RequestedBy:  userID,
		ApprovedBy:   &userID,
		CreatedAt:    now,
		CompletedAt:  &now,
	}

	if err := s.cashierRepo.CreateRefund(ctx, ref); err != nil {
		return nil, err
	}

	// If there is an active shift, record REFUND transaction in shift ledger
	shift, _ := s.cashierRepo.GetActiveShiftByRestaurant(ctx, restaurantID)
	if shift != nil {
		tx := &CashierShiftTransaction{
			ID:        "cst_" + uuid.New().String()[:8],
			ShiftID:   shift.ID,
			OrderID:   &orderID,
			PaymentID: &p.ID,
			Type:      TxTypeRefund,
			Amount:    req.Amount,
			Metadata: map[string]interface{}{
				"reason": req.Reason,
			},
			CreatedAt: now,
		}
		_ = s.cashierRepo.CreateTransaction(ctx, tx)

		if s.hub != nil {
			s.hub.Publish(fmt.Sprintf("shift:%s", shift.ID), "SHIFT_TRANSACTION_ADDED", tx)
		}
	}

	// Audit Log
	_ = s.cashierRepo.CreateAuditLog(ctx, restaurantID, userID, "REFUND_CREATED", "refunds", ref.ID, map[string]interface{}{
		"order_id": orderID,
		"amount":   req.Amount,
		"reason":   req.Reason,
	})

	return ref, nil
}

func (s *service) CreateVoid(ctx context.Context, restaurantID, orderID, userID string, req CreateVoidRequest) error {
	o, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return err
	}
	if o.Status == order.StatusCompleted {
		return ErrOrderNotEligibleVoid
	}

	// Cancel order
	if err := s.orderRepo.UpdateStatus(ctx, orderID, order.StatusCancelled, &userID); err != nil {
		return err
	}

	now := time.Now()

	// If there is an active shift, record VOID transaction in shift ledger
	shift, _ := s.cashierRepo.GetActiveShiftByRestaurant(ctx, restaurantID)
	if shift != nil {
		tx := &CashierShiftTransaction{
			ID:        "cst_" + uuid.New().String()[:8],
			ShiftID:   shift.ID,
			OrderID:   &orderID,
			Type:      TxTypeVoid,
			Amount:    o.Total,
			Metadata: map[string]interface{}{
				"reason": req.Reason,
			},
			CreatedAt: now,
		}
		_ = s.cashierRepo.CreateTransaction(ctx, tx)

		if s.hub != nil {
			s.hub.Publish(fmt.Sprintf("shift:%s", shift.ID), "SHIFT_TRANSACTION_ADDED", tx)
		}
	}

	// Audit Log
	_ = s.cashierRepo.CreateAuditLog(ctx, restaurantID, userID, "ORDER_VOIDED", "orders", orderID, map[string]interface{}{
		"reason": req.Reason,
		"amount": o.Total,
	})

	return nil
}

// ----------------- REPORTS -----------------

func (s *service) ListReports(ctx context.Context, restaurantID string, filters ShiftFilter) ([]CashierShift, error) {
	return s.cashierRepo.ListShifts(ctx, restaurantID, filters)
}

func (s *service) GetReportDetail(ctx context.Context, shiftID string) (*ShiftReportDetail, error) {
	shift, err := s.cashierRepo.GetShiftByID(ctx, shiftID)
	if err != nil {
		return nil, err
	}

	summary, err := s.cashierRepo.GetShiftSummary(ctx, shiftID)
	if err != nil {
		return nil, err
	}

	closing, _ := s.cashierRepo.GetClosingByShiftID(ctx, shiftID)
	txs, _, _ := s.cashierRepo.ListTransactions(ctx, shiftID, nil, 1, 200)

	return &ShiftReportDetail{
		Shift:        *shift,
		Summary:      *summary,
		Closing:      closing,
		Transactions: txs,
	}, nil
}
