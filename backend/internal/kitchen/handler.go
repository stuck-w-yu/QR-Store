package kitchen

import (
	"errors"

	"github.com/gin-gonic/gin"
	"qr-store/backend/internal/auth"
	"qr-store/backend/internal/order"
	"qr-store/backend/pkg/response"
)

type Handler struct {
	orderService order.Service
	authService  auth.Service
}

func NewHandler(orderService order.Service, authService auth.Service) *Handler {
	return &Handler{
		orderService: orderService,
		authService:  authService,
	}
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	kitch := r.Group("/kitchen", auth.AuthMiddleware(h.authService), auth.RequireRoles(auth.RoleKitchen, auth.RoleAdmin, auth.RoleOwner))
	{
		kitch.GET("/orders", h.ListKitchenOrders)
		kitch.POST("/orders/:id/accept", h.AcceptOrder)
		kitch.POST("/orders/:id/start", h.AcceptOrder) // PRD Section 28 contract
		kitch.POST("/orders/:id/ready", h.MarkReady)
		kitch.POST("/orders/:id/serve", h.ServeOrder)
		kitch.POST("/orders/:id/complete", h.CompleteOrder)
	}
}

func (h *Handler) ListKitchenOrders(c *gin.Context) {
	restoID, _ := c.Get(auth.CtxRestaurantID)
	orders, err := h.orderService.ListKitchenOrders(c.Request.Context(), restoID.(string))
	if err != nil {
		response.InternalServerError(c, "DB_ERROR", err.Error())
		return
	}
	response.OK(c, orders, "Kitchen orders retrieved")
}

func (h *Handler) AcceptOrder(c *gin.Context) {
	id := c.Param("id")
	userRole, _ := c.Get(auth.CtxRole)

	if err := h.orderService.UpdateOrderStatus(c.Request.Context(), id, order.StatusPreparing, string(userRole.(auth.Role))); err != nil {
		if errors.Is(err, order.ErrInvalidStatusOrder) {
			response.Conflict(c, "INVALID_STATE", "Order is not in CONFIRMED state")
			return
		}
		response.InternalServerError(c, "UPDATE_FAILED", err.Error())
		return
	}
	response.OK(c, gin.H{"status": order.StatusPreparing}, "Order accepted and preparing")
}

func (h *Handler) MarkReady(c *gin.Context) {
	id := c.Param("id")
	userRole, _ := c.Get(auth.CtxRole)

	if err := h.orderService.UpdateOrderStatus(c.Request.Context(), id, order.StatusReady, string(userRole.(auth.Role))); err != nil {
		if errors.Is(err, order.ErrInvalidStatusOrder) {
			response.Conflict(c, "INVALID_STATE", "Order is not in PREPARING state")
			return
		}
		response.InternalServerError(c, "UPDATE_FAILED", err.Error())
		return
	}
	response.OK(c, gin.H{"status": order.StatusReady}, "Order is ready for serving")
}

func (h *Handler) ServeOrder(c *gin.Context) {
	id := c.Param("id")
	userRole, _ := c.Get(auth.CtxRole)

	if err := h.orderService.UpdateOrderStatus(c.Request.Context(), id, order.StatusServed, string(userRole.(auth.Role))); err != nil {
		if errors.Is(err, order.ErrInvalidStatusOrder) {
			response.Conflict(c, "INVALID_STATE", "Order is not in READY state")
			return
		}
		response.InternalServerError(c, "UPDATE_FAILED", err.Error())
		return
	}
	response.OK(c, gin.H{"status": order.StatusServed}, "Order marked as served")
}

func (h *Handler) CompleteOrder(c *gin.Context) {
	id := c.Param("id")
	userRole, _ := c.Get(auth.CtxRole)

	if err := h.orderService.UpdateOrderStatus(c.Request.Context(), id, order.StatusCompleted, string(userRole.(auth.Role))); err != nil {
		if errors.Is(err, order.ErrInvalidStatusOrder) {
			response.Conflict(c, "INVALID_STATE", "Order is not in READY or SERVED state")
			return
		}
		response.InternalServerError(c, "UPDATE_FAILED", err.Error())
		return
	}
	response.OK(c, gin.H{"status": order.StatusCompleted}, "Order completed")
}
