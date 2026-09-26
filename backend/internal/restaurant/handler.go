package restaurant

import (
	"errors"

	"github.com/gin-gonic/gin"
	"qr-store/backend/internal/auth"
	"qr-store/backend/pkg/response"
)

type Handler struct {
	repo        Repository
	authService auth.Service
}

func NewHandler(repo Repository, authService auth.Service) *Handler {
	return &Handler{repo: repo, authService: authService}
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	resto := r.Group("/restaurants", auth.AuthMiddleware(h.authService))
	{
		resto.GET("/:id", h.GetByID)
		resto.PATCH("/:id", auth.RequireRoles(auth.RoleOwner, auth.RoleAdmin), h.Update)
	}

	// Public restaurant detail by ID or slug
	r.GET("/public/restaurants/:restaurant_id", h.GetPublicBySlug)
}

func (h *Handler) GetByID(c *gin.Context) {
	id := c.Param("id")
	tenantID, _ := c.Get(auth.CtxRestaurantID)
	if tenantID != id {
		response.Forbidden(c, "TENANT_MISMATCH", "Cannot access other restaurants")
		return
	}

	resto, err := h.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, ErrRestaurantNotFound) {
			response.NotFound(c, "RESTAURANT_NOT_FOUND", "Restaurant not found")
			return
		}
		response.InternalServerError(c, "DB_ERROR", err.Error())
		return
	}

	response.OK(c, resto, "Restaurant retrieved successfully")
}

func (h *Handler) GetPublicBySlug(c *gin.Context) {
	idOrSlug := c.Param("restaurant_id")
	resto, err := h.repo.GetByID(c.Request.Context(), idOrSlug)
	if err != nil {
		resto, err = h.repo.GetBySlug(c.Request.Context(), idOrSlug)
	}
	if err != nil {
		if errors.Is(err, ErrRestaurantNotFound) {
			response.NotFound(c, "RESTAURANT_NOT_FOUND", "Restaurant not found")
			return
		}
		response.InternalServerError(c, "DB_ERROR", err.Error())
		return
	}

	response.OK(c, resto, "Restaurant retrieved")
}

func (h *Handler) Update(c *gin.Context) {
	id := c.Param("id")
	tenantID, _ := c.Get(auth.CtxRestaurantID)
	if tenantID != id {
		response.Forbidden(c, "TENANT_MISMATCH", "Cannot modify other restaurants")
		return
	}

	var req UpdateRestaurantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_REQUEST", err.Error())
		return
	}

	updated, err := h.repo.Update(c.Request.Context(), id, req)
	if err != nil {
		response.InternalServerError(c, "UPDATE_FAILED", err.Error())
		return
	}

	response.OK(c, updated, "Restaurant updated successfully")
}
