package cashier

import (
	"errors"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"qr-store/backend/internal/auth"
	"qr-store/backend/pkg/response"
)

type Handler struct {
	service     Service
	authService auth.Service
}

func NewHandler(service Service, authService auth.Service) *Handler {
	return &Handler{
		service:     service,
		authService: authService,
	}
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	// Protected with JWT Auth Middleware
	authMW := auth.AuthMiddleware(h.authService)

	// 1. Registers
	regGroup := r.Group("/registers", authMW)
	{
		regGroup.GET("", auth.RequireRoles(auth.RoleOwner, auth.RoleAdmin, auth.RoleCashier), h.ListRegisters)
		regGroup.POST("", auth.RequireRoles(auth.RoleOwner, auth.RoleAdmin), h.CreateRegister)
		regGroup.GET("/:id", auth.RequireRoles(auth.RoleOwner, auth.RoleAdmin, auth.RoleCashier), h.GetRegister)
		regGroup.PATCH("/:id", auth.RequireRoles(auth.RoleOwner, auth.RoleAdmin), h.UpdateRegister)
		regGroup.DELETE("/:id", auth.RequireRoles(auth.RoleOwner, auth.RoleAdmin), h.DeleteRegister)
	}

	// 2. Cashier Shifts
	shiftGroup := r.Group("/cashier/shifts", authMW)
	{
		shiftGroup.GET("/current", auth.RequireRoles(auth.RoleOwner, auth.RoleAdmin, auth.RoleCashier), h.GetCurrentShift)
		shiftGroup.POST("", auth.RequireRoles(auth.RoleOwner, auth.RoleAdmin, auth.RoleCashier), h.OpenShift)
		shiftGroup.GET("/:id", auth.RequireRoles(auth.RoleOwner, auth.RoleAdmin, auth.RoleCashier), h.GetShiftByID)
		shiftGroup.GET("/:id/transactions", auth.RequireRoles(auth.RoleOwner, auth.RoleAdmin, auth.RoleCashier), h.GetShiftTransactions)
		shiftGroup.GET("/:id/summary", auth.RequireRoles(auth.RoleOwner, auth.RoleAdmin, auth.RoleCashier), h.GetShiftSummary)
		shiftGroup.POST("/:id/close", auth.RequireRoles(auth.RoleOwner, auth.RoleAdmin, auth.RoleCashier), h.CloseShift)
	}

	// 3. Refund & Void
	orderGroup := r.Group("/orders", authMW)
	{
		orderGroup.POST("/:id/refund", auth.RequireRoles(auth.RoleOwner, auth.RoleAdmin, auth.RoleCashier), h.CreateRefund)
		orderGroup.POST("/:id/void", auth.RequireRoles(auth.RoleOwner, auth.RoleAdmin, auth.RoleCashier), h.CreateVoid)
	}

	// 4. Reports
	reportGroup := r.Group("/reports/shifts", authMW)
	{
		reportGroup.GET("", auth.RequireRoles(auth.RoleOwner, auth.RoleAdmin), h.ListReports)
		reportGroup.GET("/:id", auth.RequireRoles(auth.RoleOwner, auth.RoleAdmin), h.GetReportDetail)
	}
}

// ----------------- REGISTERS HANDLERS -----------------

func (h *Handler) ListRegisters(c *gin.Context) {
	restoID, _ := c.Get(auth.CtxRestaurantID)
	list, err := h.service.ListRegisters(c.Request.Context(), restoID.(string))
	if err != nil {
		response.InternalServerError(c, "DB_ERROR", err.Error())
		return
	}
	response.OK(c, list, "Registers retrieved successfully")
}

func (h *Handler) GetRegister(c *gin.Context) {
	id := c.Param("id")
	reg, err := h.service.GetRegister(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, ErrRegisterNotFound) {
			response.NotFound(c, "REGISTER_NOT_FOUND", "Register not found")
			return
		}
		response.InternalServerError(c, "DB_ERROR", err.Error())
		return
	}
	response.OK(c, reg, "Register retrieved successfully")
}

func (h *Handler) CreateRegister(c *gin.Context) {
	restoID, _ := c.Get(auth.CtxRestaurantID)
	var req CreateRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_REQUEST", err.Error())
		return
	}

	reg, err := h.service.CreateRegister(c.Request.Context(), restoID.(string), req)
	if err != nil {
		response.InternalServerError(c, "CREATE_REGISTER_FAILED", err.Error())
		return
	}
	response.Created(c, reg, "Register created successfully")
}

func (h *Handler) UpdateRegister(c *gin.Context) {
	id := c.Param("id")
	var req UpdateRegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_REQUEST", err.Error())
		return
	}

	reg, err := h.service.UpdateRegister(c.Request.Context(), id, req)
	if err != nil {
		if errors.Is(err, ErrRegisterNotFound) {
			response.NotFound(c, "REGISTER_NOT_FOUND", "Register not found")
			return
		}
		response.InternalServerError(c, "UPDATE_REGISTER_FAILED", err.Error())
		return
	}
	response.OK(c, reg, "Register updated successfully")
}

func (h *Handler) DeleteRegister(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.DeleteRegister(c.Request.Context(), id); err != nil {
		response.InternalServerError(c, "DELETE_REGISTER_FAILED", err.Error())
		return
	}
	response.OK(c, gin.H{"deleted": true}, "Register deleted successfully")
}

// ----------------- SHIFTS HANDLERS -----------------

func (h *Handler) GetCurrentShift(c *gin.Context) {
	restoID, _ := c.Get(auth.CtxRestaurantID)
	cashierID, _ := c.Get(auth.CtxUserID)

	shift, err := h.service.GetCurrentShift(c.Request.Context(), restoID.(string), cashierID.(string))
	if err != nil {
		response.InternalServerError(c, "DB_ERROR", err.Error())
		return
	}
	if shift == nil {
		response.OK(c, nil, "No active shift found")
		return
	}
	response.OK(c, shift, "Active shift retrieved successfully")
}

func (h *Handler) OpenShift(c *gin.Context) {
	restoID, _ := c.Get(auth.CtxRestaurantID)
	cashierID, _ := c.Get(auth.CtxUserID)

	var req OpenShiftRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_REQUEST", err.Error())
		return
	}

	shift, err := h.service.OpenShift(c.Request.Context(), restoID.(string), cashierID.(string), req)
	if err != nil {
		if errors.Is(err, ErrActiveShiftExists) {
			response.Conflict(c, "SHIFT_ALREADY_ACTIVE", "An active shift already exists on this register or for this cashier")
			return
		}
		if errors.Is(err, ErrRegisterNotFound) {
			response.NotFound(c, "REGISTER_NOT_FOUND", "Register not found")
			return
		}
		if errors.Is(err, ErrOpeningBalanceNegative) {
			response.BadRequest(c, "INVALID_BALANCE", "Opening balance cannot be negative")
			return
		}
		response.InternalServerError(c, "OPEN_SHIFT_FAILED", err.Error())
		return
	}
	response.Created(c, shift, "Shift opened successfully")
}

func (h *Handler) GetShiftByID(c *gin.Context) {
	id := c.Param("id")
	shift, err := h.service.GetShiftByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, ErrShiftNotFound) {
			response.NotFound(c, "SHIFT_NOT_FOUND", "Shift not found")
			return
		}
		response.InternalServerError(c, "DB_ERROR", err.Error())
		return
	}
	response.OK(c, shift, "Shift retrieved successfully")
}

func (h *Handler) GetShiftTransactions(c *gin.Context) {
	shiftID := c.Param("id")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	var txType *TransactionType
	if t := c.Query("type"); t != "" {
		tt := TransactionType(t)
		txType = &tt
	}

	txs, total, err := h.service.GetShiftTransactions(c.Request.Context(), shiftID, txType, page, limit)
	if err != nil {
		response.InternalServerError(c, "DB_ERROR", err.Error())
		return
	}

	response.OK(c, gin.H{
		"transactions": txs,
		"total":        total,
		"page":         page,
		"limit":        limit,
	}, "Shift transactions retrieved successfully")
}

func (h *Handler) GetShiftSummary(c *gin.Context) {
	shiftID := c.Param("id")
	summary, err := h.service.GetShiftSummary(c.Request.Context(), shiftID)
	if err != nil {
		if errors.Is(err, ErrShiftNotFound) {
			response.NotFound(c, "SHIFT_NOT_FOUND", "Shift not found")
			return
		}
		response.InternalServerError(c, "SUMMARY_FAILED", err.Error())
		return
	}
	response.OK(c, summary, "Shift summary calculated successfully")
}

func (h *Handler) CloseShift(c *gin.Context) {
	shiftID := c.Param("id")
	cashierID, _ := c.Get(auth.CtxUserID)

	var req CloseShiftRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_REQUEST", err.Error())
		return
	}

	closing, err := h.service.CloseShift(c.Request.Context(), shiftID, cashierID.(string), req)
	if err != nil {
		if errors.Is(err, ErrShiftAlreadyClosed) {
			response.Conflict(c, "SHIFT_ALREADY_CLOSED", "Shift is already closed")
			return
		}
		if errors.Is(err, ErrShiftNotFound) {
			response.NotFound(c, "SHIFT_NOT_FOUND", "Shift not found")
			return
		}
		response.InternalServerError(c, "CLOSE_SHIFT_FAILED", err.Error())
		return
	}
	response.OK(c, closing, "Shift closed successfully")
}

// ----------------- REFUND & VOID HANDLERS -----------------

func (h *Handler) CreateRefund(c *gin.Context) {
	orderID := c.Param("id")
	restoID, _ := c.Get(auth.CtxRestaurantID)
	userID, _ := c.Get(auth.CtxUserID)

	var req CreateRefundRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_REQUEST", err.Error())
		return
	}

	ref, err := h.service.CreateRefund(c.Request.Context(), restoID.(string), orderID, userID.(string), req)
	if err != nil {
		if errors.Is(err, ErrOrderNotEligibleRefund) {
			response.Conflict(c, "INELIGIBLE_FOR_REFUND", err.Error())
			return
		}
		if errors.Is(err, ErrRefundAmountExceeds) {
			response.BadRequest(c, "INVALID_AMOUNT", err.Error())
			return
		}
		response.InternalServerError(c, "REFUND_FAILED", err.Error())
		return
	}
	response.Created(c, ref, "Refund processed successfully")
}

func (h *Handler) CreateVoid(c *gin.Context) {
	orderID := c.Param("id")
	restoID, _ := c.Get(auth.CtxRestaurantID)
	userID, _ := c.Get(auth.CtxUserID)

	var req CreateVoidRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_REQUEST", err.Error())
		return
	}

	if err := h.service.CreateVoid(c.Request.Context(), restoID.(string), orderID, userID.(string), req); err != nil {
		if errors.Is(err, ErrOrderNotEligibleVoid) {
			response.Conflict(c, "INELIGIBLE_FOR_VOID", err.Error())
			return
		}
		response.InternalServerError(c, "VOID_FAILED", err.Error())
		return
	}
	response.OK(c, gin.H{"voided": true}, "Order voided successfully")
}

// ----------------- REPORTS HANDLERS -----------------

func (h *Handler) ListReports(c *gin.Context) {
	restoID, _ := c.Get(auth.CtxRestaurantID)

	var filters ShiftFilter
	if cash := c.Query("cashier_id"); cash != "" {
		filters.CashierID = &cash
	}
	if reg := c.Query("register_id"); reg != "" {
		filters.RegisterID = &reg
	}
	if st := c.Query("status"); st != "" {
		s := ShiftStatus(st)
		filters.Status = &s
	}
	if start := c.Query("start_date"); start != "" {
		if t, err := time.Parse(time.RFC3339, start); err == nil {
			filters.StartDate = &t
		}
	}
	if end := c.Query("end_date"); end != "" {
		if t, err := time.Parse(time.RFC3339, end); err == nil {
			filters.EndDate = &t
		}
	}
	filters.Limit, _ = strconv.Atoi(c.DefaultQuery("limit", "50"))
	filters.Offset, _ = strconv.Atoi(c.DefaultQuery("offset", "0"))

	shifts, err := h.service.ListReports(c.Request.Context(), restoID.(string), filters)
	if err != nil {
		response.InternalServerError(c, "REPORTS_FAILED", err.Error())
		return
	}
	response.OK(c, shifts, "Shift reports retrieved successfully")
}

func (h *Handler) GetReportDetail(c *gin.Context) {
	shiftID := c.Param("id")
	report, err := h.service.GetReportDetail(c.Request.Context(), shiftID)
	if err != nil {
		if errors.Is(err, ErrShiftNotFound) {
			response.NotFound(c, "SHIFT_NOT_FOUND", "Shift not found")
			return
		}
		response.InternalServerError(c, "REPORT_DETAIL_FAILED", err.Error())
		return
	}
	response.OK(c, report, "Shift report detail retrieved successfully")
}
