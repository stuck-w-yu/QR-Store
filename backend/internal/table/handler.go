package table

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"qr-store/backend/internal/auth"
	"qr-store/backend/pkg/response"
)

type Handler struct {
	service     Service
	authService auth.Service
}

func NewHandler(service Service, authService auth.Service) *Handler {
	return &Handler{service: service, authService: authService}
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	// Public endpoint for customer scanning QR and waiter call (PRD Section 17 & 28)
	r.GET("/public/tables/:qr_token", h.GetPublicTableInfo)
	r.POST("/public/service-requests", h.CreateServiceRequest)

	// Staff management endpoints
	tables := r.Group("/tables", auth.AuthMiddleware(h.authService))
	{
		tables.GET("", h.ListTables)
		tables.POST("", auth.RequireRoles(auth.RoleOwner, auth.RoleAdmin), h.CreateTable)
		tables.GET("/:id", h.GetTable)
		tables.PATCH("/:id", auth.RequireRoles(auth.RoleOwner, auth.RoleAdmin), h.UpdateTable)
		tables.DELETE("/:id", auth.RequireRoles(auth.RoleOwner, auth.RoleAdmin), h.DeleteTable)
		tables.GET("/:id/qr", h.GetQRCodePNG)
		tables.POST("/:id/qr/regenerate", auth.RequireRoles(auth.RoleOwner, auth.RoleAdmin), h.RegenerateQR)
	}
}

func (h *Handler) GetPublicTableInfo(c *gin.Context) {
	token := c.Param("qr_token")
	info, err := h.service.GetPublicTableInfo(c.Request.Context(), token)
	if err != nil {
		if errors.Is(err, ErrTableNotFound) {
			response.NotFound(c, "TABLE_NOT_FOUND", "Table not found or inactive")
			return
		}
		response.InternalServerError(c, "DB_ERROR", err.Error())
		return
	}

	response.OK(c, info, "Table information retrieved")
}

func (h *Handler) ListTables(c *gin.Context) {
	restoID, _ := c.Get(auth.CtxRestaurantID)
	tables, err := h.service.ListTables(c.Request.Context(), restoID.(string))
	if err != nil {
		response.InternalServerError(c, "DB_ERROR", err.Error())
		return
	}
	response.OK(c, tables, "Tables retrieved successfully")
}

func (h *Handler) CreateTable(c *gin.Context) {
	restoID, _ := c.Get(auth.CtxRestaurantID)
	var req CreateTableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_REQUEST", err.Error())
		return
	}

	t, err := h.service.CreateTable(c.Request.Context(), restoID.(string), req)
	if err != nil {
		response.InternalServerError(c, "CREATE_FAILED", err.Error())
		return
	}
	response.Created(c, t, "Table created successfully")
}

func (h *Handler) GetTable(c *gin.Context) {
	restoID, _ := c.Get(auth.CtxRestaurantID)
	id := c.Param("id")

	t, err := h.service.GetTable(c.Request.Context(), id, restoID.(string))
	if err != nil {
		if errors.Is(err, ErrTableNotFound) {
			response.NotFound(c, "TABLE_NOT_FOUND", "Table not found")
			return
		}
		response.InternalServerError(c, "DB_ERROR", err.Error())
		return
	}
	response.OK(c, t, "Table retrieved successfully")
}

func (h *Handler) UpdateTable(c *gin.Context) {
	restoID, _ := c.Get(auth.CtxRestaurantID)
	id := c.Param("id")
	var req UpdateTableRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_REQUEST", err.Error())
		return
	}

	t, err := h.service.UpdateTable(c.Request.Context(), id, restoID.(string), req)
	if err != nil {
		response.InternalServerError(c, "UPDATE_FAILED", err.Error())
		return
	}
	response.OK(c, t, "Table updated successfully")
}

func (h *Handler) DeleteTable(c *gin.Context) {
	restoID, _ := c.Get(auth.CtxRestaurantID)
	id := c.Param("id")

	if err := h.service.DeleteTable(c.Request.Context(), id, restoID.(string)); err != nil {
		response.InternalServerError(c, "DELETE_FAILED", err.Error())
		return
	}
	response.OK(c, gin.H{"deleted": true}, "Table deleted successfully")
}

func (h *Handler) RegenerateQR(c *gin.Context) {
	restoID, _ := c.Get(auth.CtxRestaurantID)
	id := c.Param("id")

	newToken, err := h.service.RegenerateQRToken(c.Request.Context(), id, restoID.(string))
	if err != nil {
		response.InternalServerError(c, "REGENERATE_FAILED", err.Error())
		return
	}
	response.OK(c, gin.H{"qr_token": newToken}, "QR Token regenerated successfully")
}

func (h *Handler) GetQRCodePNG(c *gin.Context) {
	restoID, _ := c.Get(auth.CtxRestaurantID)
	id := c.Param("id")

	baseURL := c.Query("base_url")
	if baseURL == "" {
		baseURL = "http://localhost:5173"
	}

	pngBytes, err := h.service.GenerateQRCodePNG(c.Request.Context(), id, restoID.(string), baseURL)
	if err != nil {
		response.InternalServerError(c, "QR_GENERATE_FAILED", err.Error())
		return
	}

	c.Data(http.StatusOK, "image/png", pngBytes)
}

func (h *Handler) CreateServiceRequest(c *gin.Context) {
	var req CreateServiceRequestInput
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_REQUEST", err.Error())
		return
	}

	sr, err := h.service.CreateServiceRequest(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, ErrTableNotFound) {
			response.NotFound(c, "TABLE_NOT_FOUND", "Table not found or inactive")
			return
		}
		response.InternalServerError(c, "SERVICE_REQUEST_FAILED", err.Error())
		return
	}

	response.Created(c, sr, "Service request submitted successfully")
}
