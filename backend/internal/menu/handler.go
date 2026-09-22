package menu

import (
	"errors"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"qr-store/backend/internal/auth"
	"qr-store/backend/pkg/response"
)

type CreateCategoryRequest struct {
	Name      string `json:"name" binding:"required"`
	SortOrder int    `json:"sort_order"`
}

type UpdateCategoryRequest struct {
	Name      string `json:"name" binding:"required"`
	SortOrder int    `json:"sort_order"`
	Status    string `json:"status" binding:"required"`
}

type CreateMenuRequest struct {
	CategoryID  *string `json:"category_id"`
	Name        string  `json:"name" binding:"required"`
	Description *string `json:"description"`
	Price       int64   `json:"price" binding:"required,gte=0"`
	ImageURL    *string `json:"image_url"`
	Available   bool    `json:"available"`
	SortOrder   int     `json:"sort_order"`
}

type UpdateMenuRequest struct {
	CategoryID  *string `json:"category_id"`
	Name        string  `json:"name" binding:"required"`
	Description *string `json:"description"`
	Price       int64   `json:"price" binding:"required,gte=0"`
	ImageURL    *string `json:"image_url"`
	Available   bool    `json:"available"`
	SortOrder   int     `json:"sort_order"`
}

type UpdateAvailabilityRequest struct {
	Available bool `json:"available"`
}

type CreateModifierRequest struct {
	MenuID   *string `json:"menu_id"`
	Name     string  `json:"name" binding:"required"`
	Type     string  `json:"type" binding:"required"` // SINGLE, MULTIPLE
	Required bool    `json:"required"`
	Options  []struct {
		Name            string `json:"name" binding:"required"`
		AdditionalPrice int64  `json:"additional_price"`
	} `json:"options" binding:"required,min=1"`
}

type Handler struct {
	repo        Repository
	authService auth.Service
}

func NewHandler(repo Repository, authService auth.Service) *Handler {
	return &Handler{repo: repo, authService: authService}
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	// Public Catalog
	r.GET("/public/restaurants/:restaurant_id/menu", h.GetPublicMenu)

	// Admin / Owner Routes
	catGroup := r.Group("/categories", auth.AuthMiddleware(h.authService))
	{
		catGroup.GET("", h.ListCategories)
		catGroup.POST("", auth.RequireRoles(auth.RoleOwner, auth.RoleAdmin), h.CreateCategory)
		catGroup.PATCH("/:id", auth.RequireRoles(auth.RoleOwner, auth.RoleAdmin), h.UpdateCategory)
		catGroup.DELETE("/:id", auth.RequireRoles(auth.RoleOwner, auth.RoleAdmin), h.DeleteCategory)
	}

	menuGroup := r.Group("/menus", auth.AuthMiddleware(h.authService))
	{
		menuGroup.GET("", h.ListMenus)
		menuGroup.POST("", auth.RequireRoles(auth.RoleOwner, auth.RoleAdmin), h.CreateMenu)
		menuGroup.GET("/:id", h.GetMenu)
		menuGroup.PATCH("/:id", auth.RequireRoles(auth.RoleOwner, auth.RoleAdmin), h.UpdateMenu)
		menuGroup.PATCH("/:id/availability", auth.RequireRoles(auth.RoleOwner, auth.RoleAdmin, auth.RoleCashier, auth.RoleKitchen), h.UpdateAvailability)
		menuGroup.DELETE("/:id", auth.RequireRoles(auth.RoleOwner, auth.RoleAdmin), h.DeleteMenu)
		menuGroup.POST("/modifiers", auth.RequireRoles(auth.RoleOwner, auth.RoleAdmin), h.CreateModifier)
		menuGroup.DELETE("/modifiers/:id", auth.RequireRoles(auth.RoleOwner, auth.RoleAdmin), h.DeleteModifier)
	}
}

func (h *Handler) GetPublicMenu(c *gin.Context) {
	restoID := c.Param("restaurant_id")
	catalog, err := h.repo.GetPublicCatalog(c.Request.Context(), restoID)
	if err != nil {
		response.InternalServerError(c, "DB_ERROR", err.Error())
		return
	}
	response.OK(c, catalog, "Menu catalog retrieved successfully")
}

// Categories
func (h *Handler) ListCategories(c *gin.Context) {
	restoID, _ := c.Get(auth.CtxRestaurantID)
	categories, err := h.repo.ListCategories(c.Request.Context(), restoID.(string))
	if err != nil {
		response.InternalServerError(c, "DB_ERROR", err.Error())
		return
	}
	response.OK(c, categories, "Categories retrieved")
}

func (h *Handler) CreateCategory(c *gin.Context) {
	restoID, _ := c.Get(auth.CtxRestaurantID)
	var req CreateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_REQUEST", err.Error())
		return
	}

	now := time.Now()
	cat := &Category{
		ID:           "cat_" + uuid.New().String()[:8],
		RestaurantID: restoID.(string),
		Name:         req.Name,
		SortOrder:    req.SortOrder,
		Status:       "ACTIVE",
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := h.repo.CreateCategory(c.Request.Context(), cat); err != nil {
		response.InternalServerError(c, "CREATE_FAILED", err.Error())
		return
	}
	response.Created(c, cat, "Category created successfully")
}

func (h *Handler) UpdateCategory(c *gin.Context) {
	restoID, _ := c.Get(auth.CtxRestaurantID)
	id := c.Param("id")

	var req UpdateCategoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_REQUEST", err.Error())
		return
	}

	cat, err := h.repo.GetCategoryByID(c.Request.Context(), id, restoID.(string))
	if err != nil {
		if errors.Is(err, ErrCategoryNotFound) {
			response.NotFound(c, "CATEGORY_NOT_FOUND", "Category not found")
			return
		}
		response.InternalServerError(c, "DB_ERROR", err.Error())
		return
	}

	cat.Name = req.Name
	cat.SortOrder = req.SortOrder
	cat.Status = req.Status

	if err := h.repo.UpdateCategory(c.Request.Context(), cat); err != nil {
		response.InternalServerError(c, "UPDATE_FAILED", err.Error())
		return
	}
	response.OK(c, cat, "Category updated successfully")
}

func (h *Handler) DeleteCategory(c *gin.Context) {
	restoID, _ := c.Get(auth.CtxRestaurantID)
	id := c.Param("id")

	if err := h.repo.DeleteCategory(c.Request.Context(), id, restoID.(string)); err != nil {
		response.InternalServerError(c, "DELETE_FAILED", err.Error())
		return
	}
	response.OK(c, gin.H{"deleted": true}, "Category deleted successfully")
}

// Menus
func (h *Handler) ListMenus(c *gin.Context) {
	restoID, _ := c.Get(auth.CtxRestaurantID)
	var catID *string
	if q := c.Query("category_id"); q != "" {
		catID = &q
	}

	menus, err := h.repo.ListMenus(c.Request.Context(), restoID.(string), catID)
	if err != nil {
		response.InternalServerError(c, "DB_ERROR", err.Error())
		return
	}
	response.OK(c, menus, "Menus retrieved")
}

func (h *Handler) GetMenu(c *gin.Context) {
	restoID, _ := c.Get(auth.CtxRestaurantID)
	id := c.Param("id")

	m, err := h.repo.GetMenuByID(c.Request.Context(), id, restoID.(string))
	if err != nil {
		if errors.Is(err, ErrMenuNotFound) {
			response.NotFound(c, "MENU_NOT_FOUND", "Menu item not found")
			return
		}
		response.InternalServerError(c, "DB_ERROR", err.Error())
		return
	}
	response.OK(c, m, "Menu item retrieved")
}

func (h *Handler) CreateMenu(c *gin.Context) {
	restoID, _ := c.Get(auth.CtxRestaurantID)
	var req CreateMenuRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_REQUEST", err.Error())
		return
	}

	now := time.Now()
	m := &Menu{
		ID:           "menu_" + uuid.New().String()[:8],
		RestaurantID: restoID.(string),
		CategoryID:   req.CategoryID,
		Name:         req.Name,
		Description:  req.Description,
		Price:        req.Price,
		ImageURL:     req.ImageURL,
		Available:    req.Available,
		SortOrder:    req.SortOrder,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := h.repo.CreateMenu(c.Request.Context(), m); err != nil {
		response.InternalServerError(c, "CREATE_FAILED", err.Error())
		return
	}
	response.Created(c, m, "Menu item created successfully")
}

func (h *Handler) UpdateMenu(c *gin.Context) {
	restoID, _ := c.Get(auth.CtxRestaurantID)
	id := c.Param("id")

	var req UpdateMenuRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_REQUEST", err.Error())
		return
	}

	m, err := h.repo.GetMenuByID(c.Request.Context(), id, restoID.(string))
	if err != nil {
		if errors.Is(err, ErrMenuNotFound) {
			response.NotFound(c, "MENU_NOT_FOUND", "Menu item not found")
			return
		}
		response.InternalServerError(c, "DB_ERROR", err.Error())
		return
	}

	m.CategoryID = req.CategoryID
	m.Name = req.Name
	m.Description = req.Description
	m.Price = req.Price
	m.ImageURL = req.ImageURL
	m.Available = req.Available
	m.SortOrder = req.SortOrder

	if err := h.repo.UpdateMenu(c.Request.Context(), m); err != nil {
		response.InternalServerError(c, "UPDATE_FAILED", err.Error())
		return
	}
	response.OK(c, m, "Menu item updated successfully")
}

func (h *Handler) UpdateAvailability(c *gin.Context) {
	restoID, _ := c.Get(auth.CtxRestaurantID)
	id := c.Param("id")

	var req UpdateAvailabilityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_REQUEST", err.Error())
		return
	}

	if err := h.repo.UpdateMenuAvailability(c.Request.Context(), id, restoID.(string), req.Available); err != nil {
		response.InternalServerError(c, "UPDATE_FAILED", err.Error())
		return
	}
	response.OK(c, gin.H{"available": req.Available}, "Menu availability updated")
}

func (h *Handler) DeleteMenu(c *gin.Context) {
	restoID, _ := c.Get(auth.CtxRestaurantID)
	id := c.Param("id")

	if err := h.repo.DeleteMenu(c.Request.Context(), id, restoID.(string)); err != nil {
		response.InternalServerError(c, "DELETE_FAILED", err.Error())
		return
	}
	response.OK(c, gin.H{"deleted": true}, "Menu item deleted successfully")
}

func (h *Handler) CreateModifier(c *gin.Context) {
	restoID, _ := c.Get(auth.CtxRestaurantID)
	var req CreateModifierRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_REQUEST", err.Error())
		return
	}

	now := time.Now()
	mod := &Modifier{
		ID:           "mod_" + uuid.New().String()[:8],
		RestaurantID: restoID.(string),
		MenuID:       req.MenuID,
		Name:         req.Name,
		Type:         req.Type,
		Required:     req.Required,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	for _, opt := range req.Options {
		mod.Options = append(mod.Options, ModifierOption{
			ID:              "opt_" + uuid.New().String()[:8],
			ModifierID:      mod.ID,
			Name:            opt.Name,
			AdditionalPrice: opt.AdditionalPrice,
			CreatedAt:       now,
			UpdatedAt:       now,
		})
	}

	if err := h.repo.CreateModifier(c.Request.Context(), mod); err != nil {
		response.InternalServerError(c, "CREATE_MODIFIER_FAILED", err.Error())
		return
	}
	response.Created(c, mod, "Modifier created successfully")
}

func (h *Handler) DeleteModifier(c *gin.Context) {
	restoID, _ := c.Get(auth.CtxRestaurantID)
	id := c.Param("id")

	if err := h.repo.DeleteModifier(c.Request.Context(), id, restoID.(string)); err != nil {
		response.InternalServerError(c, "DELETE_FAILED", err.Error())
		return
	}
	response.OK(c, gin.H{"deleted": true}, "Modifier deleted successfully")
}
