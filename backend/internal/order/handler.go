package order

import (
	"errors"

	"github.com/gin-gonic/gin"
	"qr-store/backend/internal/auth"
	"qr-store/backend/pkg/response"
)

type UpdateStatusRequest struct {
	Status Status `json:"status" binding:"required"`
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

	// Staff routes
	orderGroup := r.Group("/orders", auth.AuthMiddleware(h.authService))
	{
		orderGroup.GET("", h.ListOrders)
		orderGroup.GET("/:id", h.GetOrder)
		orderGroup.PATCH("/:id/status", auth.RequireRoles(auth.RoleOwner, auth.RoleAdmin, auth.RoleCashier), h.UpdateStatus)
		orderGroup.POST("/:id/cancel", auth.RequireRoles(auth.RoleOwner, auth.RoleAdmin, auth.RoleCashier), h.CancelOrder)
		orderGroup.GET("/analytics/today", auth.RequireRoles(auth.RoleOwner, auth.RoleAdmin, auth.RoleCashier), h.GetTodayAnalytics)
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
			response.Conflict(c, "ITEM_UNAVAILABLE", err.Error())
			return
		}
		response.InternalServerError(c, "CREATE_ORDER_FAILED", err.Error())
		return
	}

	response.Created(c, order, "Order created successfully")
}

func (h *Handler) GetPublicOrder(c *gin.Context) {
	id := c.Param("id")
	order, err := h.service.GetOrderByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, ErrOrderNotFound) {
			response.NotFound(c, "ORDER_NOT_FOUND", "Order not found")
			return
		}
		response.InternalServerError(c, "DB_ERROR", err.Error())
		return
	}
	response.OK(c, order, "Order retrieved successfully")
}

func (h *Handler) ListOrders(c *gin.Context) {
	restoID, _ := c.Get(auth.CtxRestaurantID)
	var status *Status
	if s := c.Query("status"); s != "" {
		st := Status(s)
		status = &st
	}

	orders, err := h.service.ListOrders(c.Request.Context(), restoID.(string), status)
	if err != nil {
		response.InternalServerError(c, "DB_ERROR", err.Error())
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
		response.InternalServerError(c, "DB_ERROR", err.Error())
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
		response.InternalServerError(c, "UPDATE_STATUS_FAILED", err.Error())
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
		response.InternalServerError(c, "CANCEL_FAILED", err.Error())
		return
	}

	response.OK(c, gin.H{"cancelled": true}, "Order cancelled successfully")
}

func (h *Handler) GetTodayAnalytics(c *gin.Context) {
	restoID, _ := c.Get(auth.CtxRestaurantID)
	stats, err := h.service.GetTodayAnalytics(c.Request.Context(), restoID.(string))
	if err != nil {
		response.InternalServerError(c, "ANALYTICS_FAILED", err.Error())
		return
	}
	response.OK(c, stats, "Today's analytics retrieved successfully")
}
