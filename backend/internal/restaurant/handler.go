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

	// Superadmin endpoints: accessible only by SUPERADMIN (IT)
	super := r.Group("/superadmin", auth.AuthMiddleware(h.authService), auth.RequireRoles(auth.RoleSuperadmin))
	{
		super.GET("/stats", h.GetPlatformStats)
		super.GET("/restaurants", h.ListTenants)
		super.POST("/restaurants", h.OnboardTenant)
		super.PATCH("/restaurants/:id", h.UpdateTenant)
		super.DELETE("/restaurants/:id", h.DeleteTenant)
	}

	// Public restaurant detail by ID or slug
	r.GET("/public/restaurants/:restaurant_id", h.GetPublicBySlug)
}

func (h *Handler) GetByID(c *gin.Context) {
	id := c.Param("id")
	tenantID, _ := c.Get(auth.CtxRestaurantID)
	role, _ := c.Get(auth.CtxRole)
	if role != auth.RoleSuperadmin && tenantID != id {
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
	role, _ := c.Get(auth.CtxRole)
	if role != auth.RoleSuperadmin && tenantID != id {
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

// Superadmin Handlers

func (h *Handler) GetPlatformStats(c *gin.Context) {
	stats, err := h.repo.GetPlatformStats(c.Request.Context())
	if err != nil {
		response.InternalServerError(c, "DB_ERROR", err.Error())
		return
	}
	response.OK(c, stats, "Platform statistics retrieved successfully")
}

func (h *Handler) ListTenants(c *gin.Context) {
	tenants, err := h.repo.ListTenants(c.Request.Context())
	if err != nil {
		response.InternalServerError(c, "DB_ERROR", err.Error())
		return
	}
	response.OK(c, tenants, "Tenants retrieved successfully")
}

func (h *Handler) OnboardTenant(c *gin.Context) {
	var req OnboardTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_REQUEST", err.Error())
		return
	}

	hash, err := h.authService.HashPassword(req.OwnerPassword)
	if err != nil {
		response.InternalServerError(c, "HASH_FAILED", "Failed to hash password")
		return
	}

	tenant, err := h.repo.OnboardTenant(c.Request.Context(), req, hash)
	if err != nil {
		response.InternalServerError(c, "ONBOARD_FAILED", err.Error())
		return
	}

	response.Created(c, tenant, "Business owner & restaurant onboarded successfully")
}

func (h *Handler) UpdateTenant(c *gin.Context) {
	id := c.Param("id")
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

	response.OK(c, updated, "Tenant updated successfully")
}

func (h *Handler) DeleteTenant(c *gin.Context) {
	id := c.Param("id")
	if err := h.repo.Delete(c.Request.Context(), id); err != nil {
		response.InternalServerError(c, "DELETE_FAILED", err.Error())
		return
	}
	response.OK(c, nil, "Tenant deleted successfully")
}
