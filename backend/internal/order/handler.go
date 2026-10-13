package order

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"
	"qr-store/backend/internal/auth"
	"qr-store/backend/pkg/response"
)

type UpdateStatusRequest struct {
	Status Status `json:"status" binding:"required"`
}

type SubmitProofRequest struct {
	ProofURL string `json:"proof_url" binding:"required"`
}

type UpdateOrderDetailsRequest struct {
	Notes         *string        `json:"notes"`
	Status        *Status        `json:"status"`
	PaymentStatus *PaymentStatus `json:"payment_status"`
}

type Handler struct {
	service     Service
	authService auth.Service
}

func NewHandler(service Service, authService auth.Service) *Handler {
	return &Handler{service: service, authService: authService}
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	// Public customer routes
	r.POST("/public/orders", h.CreatePublicOrder)
	r.GET("/public/orders/:id", h.GetPublicOrder)
	r.GET("/public/orders/:id/status", h.GetPublicOrderStatus)
	r.POST("/public/orders/:id/proof", h.SubmitProof)

	// Staff routes
	orderGroup := r.Group("/orders", auth.AuthMiddleware(h.authService))
	{
		orderGroup.GET("", h.ListOrders)
		orderGroup.GET("/:id", h.GetOrder)
		orderGroup.PATCH("/:id/status", auth.RequireRoles(auth.RoleSuperadmin, auth.RoleOwner, auth.RoleAdmin, auth.RoleCashier), h.UpdateStatus)
		orderGroup.PATCH("/:id", auth.RequireRoles(auth.RoleSuperadmin, auth.RoleOwner, auth.RoleAdmin, auth.RoleCashier), h.UpdateOrderDetails)
		orderGroup.POST("/:id/cancel", auth.RequireRoles(auth.RoleSuperadmin, auth.RoleOwner, auth.RoleAdmin, auth.RoleCashier), h.CancelOrder)
		orderGroup.DELETE("/:id", auth.RequireRoles(auth.RoleSuperadmin, auth.RoleOwner, auth.RoleAdmin, auth.RoleCashier), h.DeleteOrder)
		orderGroup.GET("/analytics/today", auth.RequireRoles(auth.RoleSuperadmin, auth.RoleOwner, auth.RoleAdmin, auth.RoleCashier), h.GetTodayAnalytics)
	}
}

func (h *Handler) CreatePublicOrder(c *gin.Context) {
	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_REQUEST", err.Error())
		return
	}

	order, err := h.service.CreatePublicOrder(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, ErrTableInactive) || errors.Is(err, ErrRestaurantInactive) {
			response.BadRequest(c, "INVALID_TABLE_OR_RESTAURANT", err.Error())
			return
		}
		if errors.Is(err, ErrMenuUnavailable) {
			response.Conflict(c, "ITEM_UNAVAILABLE", "One or more items are unavailable")
			return
		}
		response.InternalServerError(c, "CREATE_ORDER_FAILED", "Failed to create order")
		return
	}

	response.Created(c, order, "Order created successfully")
}

func (h *Handler) GetPublicOrder(c *gin.Context) {
	id := c.Param("id")
	qrToken := publicToken(c)
	order, err := h.service.GetPublicOrder(c.Request.Context(), id, qrToken)
	if err != nil {
		if errors.Is(err, ErrOrderNotFound) {
			response.NotFound(c, "ORDER_NOT_FOUND", "Order not found")
			return
		}
		if errors.Is(err, ErrOrderAccessDenied) {
			response.Forbidden(c, "ORDER_ACCESS_DENIED", "You are not allowed to view this order")
			return
		}
		response.InternalServerError(c, "DB_ERROR", "Failed to retrieve order")
		return
	}
	response.OK(c, order, "Order retrieved successfully")
}

func (h *Handler) GetPublicOrderStatus(c *gin.Context) {
	id := c.Param("id")
	qrToken := publicToken(c)
	order, err := h.service.GetPublicOrder(c.Request.Context(), id, qrToken)
	if err != nil {
		if errors.Is(err, ErrOrderNotFound) {
			response.NotFound(c, "ORDER_NOT_FOUND", "Order not found")
			return
		}
		if errors.Is(err, ErrOrderAccessDenied) {
			response.Forbidden(c, "ORDER_ACCESS_DENIED", "You are not allowed to view this order")
			return
		}
		response.InternalServerError(c, "DB_ERROR", "Failed to retrieve order status")
		return
	}
	response.OK(c, gin.H{
		"order_id":       order.ID,
		"order_number":   order.OrderNumber,
		"status":         order.Status,
		"payment_status": order.PaymentStatus,
		"payment_method": order.PaymentMethod,
		"updated_at":     order.UpdatedAt,
		"completed_at":   order.CompletedAt,
		"cancelled_at":   order.CancelledAt,
	}, "Order status retrieved successfully")
}

// publicToken extracts the customer's QR session token from query/header.
func publicToken(c *gin.Context) string {
	if t := c.Query("qr_token"); t != "" {
		return t
	}
	if h := c.GetHeader("X-QR-Token"); h != "" {
		return h
	}
	return ""
}

func (h *Handler) ListOrders(c *gin.Context) {
	restoID, _ := c.Get(auth.CtxRestaurantID)
	var status *Status
	if s := c.Query("status"); s != "" {
		st := Status(s)
		status = &st
	}

	limit := 50
	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil {
			limit = v
		}
	}
	offset := 0
	if o := c.Query("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil && v >= 0 {
			offset = v
		}
	}

	orders, err := h.service.ListOrders(c.Request.Context(), restoID.(string), status, limit, offset)
	if err != nil {
		response.ServerError(c, "DB_ERROR", err)
		return
	}
	response.OK(c, orders, "Orders retrieved successfully")
}

func (h *Handler) GetOrder(c *gin.Context) {
	id := c.Param("id")
	order, err := h.service.GetOrderByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, ErrOrderNotFound) {
			response.NotFound(c, "ORDER_NOT_FOUND", "Order not found")
			return
		}
		response.InternalServerError(c, "DB_ERROR", "Failed to retrieve order")
		return
	}
	response.OK(c, order, "Order retrieved successfully")
}

func (h *Handler) UpdateStatus(c *gin.Context) {
	id := c.Param("id")
	userRole, _ := c.Get(auth.CtxRole)

	var req UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_REQUEST", err.Error())
		return
	}

	if err := h.service.UpdateOrderStatus(c.Request.Context(), id, req.Status, string(userRole.(auth.Role))); err != nil {
		if errors.Is(err, ErrInvalidStatusOrder) {
			response.Conflict(c, "INVALID_STATUS_TRANSITION", "This order cannot transition to the requested status")
			return
		}
		response.InternalServerError(c, "UPDATE_STATUS_FAILED", "Failed to update order status")
		return
	}

	response.OK(c, gin.H{"updated": true, "status": req.Status}, "Order status updated successfully")
}

func (h *Handler) CancelOrder(c *gin.Context) {
	id := c.Param("id")
	userRole, _ := c.Get(auth.CtxRole)

	if err := h.service.UpdateOrderStatus(c.Request.Context(), id, StatusCancelled, string(userRole.(auth.Role))); err != nil {
		if errors.Is(err, ErrInvalidStatusOrder) {
			response.Conflict(c, "CANNOT_CANCEL", "Completed or cancelled orders cannot be cancelled")
			return
		}
		response.InternalServerError(c, "CANCEL_FAILED", "Failed to cancel order")
		return
	}

	response.OK(c, gin.H{"cancelled": true}, "Order cancelled successfully")
}

func (h *Handler) UpdateOrderDetails(c *gin.Context) {
	id := c.Param("id")
	restoID, _ := c.Get(auth.CtxRestaurantID)

	var req UpdateOrderDetailsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_REQUEST", err.Error())
		return
	}

	o, err := h.service.UpdateOrderDetails(c.Request.Context(), id, restoID.(string), req.Notes, req.Status, req.PaymentStatus)
	if err != nil {
		if errors.Is(err, ErrOrderNotFound) {
			response.NotFound(c, "ORDER_NOT_FOUND", "Order not found")
			return
		}
		response.InternalServerError(c, "UPDATE_DETAILS_FAILED", "Failed to update order details")
		return
	}

	response.OK(c, o, "Order details updated successfully")
}

func (h *Handler) DeleteOrder(c *gin.Context) {
	id := c.Param("id")
	restoID, _ := c.Get(auth.CtxRestaurantID)

	if err := h.service.DeleteOrder(c.Request.Context(), id, restoID.(string)); err != nil {
		if errors.Is(err, ErrOrderNotFound) {
			response.NotFound(c, "ORDER_NOT_FOUND", "Order not found")
			return
		}
		response.InternalServerError(c, "DELETE_ORDER_FAILED", "Failed to delete order")
		return
	}

	response.OK(c, gin.H{"deleted": true}, "Order deleted successfully")
}

func (h *Handler) GetTodayAnalytics(c *gin.Context) {
	restoID, _ := c.Get(auth.CtxRestaurantID)
	stats, err := h.service.GetTodayAnalytics(c.Request.Context(), restoID.(string))
	if err != nil {
		response.InternalServerError(c, "ANALYTICS_FAILED", "Failed to load analytics")
		return
	}
	response.OK(c, stats, "Today's analytics retrieved successfully")
}

func (h *Handler) SubmitProof(c *gin.Context) {
	id := c.Param("id")
	var req SubmitProofRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_REQUEST", "Invalid request body")
		return
	}

	qrToken := publicToken(c)
	o, err := h.service.SubmitProofOfPayment(c.Request.Context(), id, qrToken, req.ProofURL)
	if err != nil {
		switch {
		case errors.Is(err, ErrOrderNotFound):
			response.NotFound(c, "ORDER_NOT_FOUND", "Order not found")
		case errors.Is(err, ErrOrderAccessDenied):
			response.Forbidden(c, "ORDER_ACCESS_DENIED", "You are not allowed to modify this order")
		case errors.Is(err, ErrInvalidProofURL):
			response.BadRequest(c, "INVALID_PROOF_URL", "Proof URL must point to an allowed image host")
		case errors.Is(err, ErrProofNotAllowed):
			response.Conflict(c, "PROOF_NOT_ALLOWED", "Proof of payment is not allowed for this order status")
		default:
			response.InternalServerError(c, "SUBMIT_PROOF_FAILED", "Failed to submit proof of payment")
		}
		return
	}

	response.OK(c, o, "Proof of payment submitted successfully")
}
