package payment

import (
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"qr-store/backend/pkg/response"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	// Public payment endpoints
	r.POST("/orders/:id/payment", h.CreatePayment)
	r.GET("/orders/:id/payment", h.GetPaymentByOrderID)
	r.GET("/payments/:id", h.GetPayment)
	r.POST("/payments/webhook", h.HandleWebhook)
}

func (h *Handler) CreatePayment(c *gin.Context) {
	orderID := c.Param("id")

	payment, err := h.service.CreatePaymentForOrder(c.Request.Context(), orderID)
	if err != nil {
		if errors.Is(err, ErrOrderAlreadyPaid) {
			response.Conflict(c, "ORDER_ALREADY_PAID", "Order is already paid or completed")
			return
		}
		response.InternalServerError(c, "PAYMENT_CREATION_FAILED", err.Error())
		return
	}
	response.Created(c, payment, "Payment initialized successfully")
}

func (h *Handler) GetPayment(c *gin.Context) {
	id := c.Param("id")
	p, err := h.service.GetPaymentByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, ErrPaymentNotFound) {
			response.NotFound(c, "PAYMENT_NOT_FOUND", "Payment record not found")
			return
		}
		response.InternalServerError(c, "DB_ERROR", err.Error())
		return
	}
	response.OK(c, p, "Payment details retrieved")
}

func (h *Handler) GetPaymentByOrderID(c *gin.Context) {
	orderID := c.Param("id")
	p, err := h.service.GetPaymentByOrderID(c.Request.Context(), orderID)
	if err != nil {
		if errors.Is(err, ErrPaymentNotFound) {
			response.NotFound(c, "PAYMENT_NOT_FOUND", "Payment record not found for this order")
			return
		}
		response.InternalServerError(c, "DB_ERROR", err.Error())
		return
	}
	response.OK(c, p, "Payment details retrieved")
}


func (h *Handler) HandleWebhook(c *gin.Context) {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		response.BadRequest(c, "INVALID_BODY", "Cannot read webhook body")
		return
	}

	result, err := h.service.ProcessWebhook(c.Request.Context(), c.Request.Header, body)
	if err != nil {
		if errors.Is(err, ErrInvalidAmount) {
			response.BadRequest(c, "INVALID_PAYMENT_AMOUNT", "Webhook amount does not match order")
			return
		}
		response.Error(c, http.StatusUnauthorized, "INVALID_WEBHOOK", err.Error())
		return
	}

	response.OK(c, gin.H{
		"event_id": result.EventID,
		"status":   result.Status,
	}, "Webhook processed successfully")
}

type SimulatePayRequest struct {
	OrderID       string `json:"order_id" binding:"required"`
	PaymentMethod string `json:"payment_method"`
}

func (h *Handler) SimulatePay(c *gin.Context) {
	var req SimulatePayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_REQUEST", err.Error())
		return
	}

	p, err := h.service.SimulatePay(c.Request.Context(), req.OrderID, req.PaymentMethod)
	if err != nil {
		if errors.Is(err, ErrOrderAlreadyPaid) {
			response.Conflict(c, "ORDER_ALREADY_PAID", "Order is already paid")
			return
		}
		response.InternalServerError(c, "SIMULATE_PAY_FAILED", err.Error())
		return
	}

	response.OK(c, p, "Simulated payment successful")
}
