package user

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"qr-store/backend/internal/auth"
	"qr-store/backend/pkg/response"
)

type CreateUserRequest struct {
	Name     string    `json:"name" binding:"required"`
	Email    string    `json:"email" binding:"required,email"`
	Password string    `json:"password" binding:"required,min=6"`
	Role     auth.Role `json:"role" binding:"required"`
}

type Handler struct {
	authRepo    auth.Repository
	authService auth.Service
}

func NewHandler(authRepo auth.Repository, authService auth.Service) *Handler {
	return &Handler{authRepo: authRepo, authService: authService}
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	users := r.Group("/users", auth.AuthMiddleware(h.authService), auth.RequireRoles(auth.RoleOwner, auth.RoleAdmin))
	{
		users.GET("", h.ListUsers)
		users.POST("", h.CreateUser)
		users.DELETE("/:id", h.DeleteUser)
	}
}

func (h *Handler) ListUsers(c *gin.Context) {
	restoID, _ := c.Get(auth.CtxRestaurantID)
	users, err := h.authRepo.ListByRestaurant(c.Request.Context(), restoID.(string))
	if err != nil {
		response.InternalServerError(c, "DB_ERROR", err.Error())
		return
	}

	var dtos []auth.UserDTO
	for _, u := range users {
		dtos = append(dtos, auth.UserDTO{
			ID:           u.ID,
			RestaurantID: u.RestaurantID,
			Name:         u.Name,
			Email:        u.Email,
			Role:         u.Role,
			Status:       u.Status,
		})
	}

	response.OK(c, dtos, "Users retrieved successfully")
}

func (h *Handler) CreateUser(c *gin.Context) {
	restoID, _ := c.Get(auth.CtxRestaurantID)
	currentRole, _ := c.Get(auth.CtxRole)

	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_REQUEST", err.Error())
		return
	}

	// Admin cannot create Owner
	if currentRole != auth.RoleOwner && req.Role == auth.RoleOwner {
		response.Forbidden(c, "FORBIDDEN", "Only Owner can create another Owner")
		return
	}

	hash, err := h.authService.HashPassword(req.Password)
	if err != nil {
		response.InternalServerError(c, "HASH_FAILED", err.Error())
		return
	}

	now := time.Now()
	u := &auth.User{
		ID:           "usr_" + uuid.New().String()[:8],
		RestaurantID: restoID.(string),
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: hash,
		Role:         req.Role,
		Status:       "ACTIVE",
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := h.authRepo.Create(c.Request.Context(), u); err != nil {
		response.InternalServerError(c, "CREATE_USER_FAILED", err.Error())
		return
	}

	dto := auth.UserDTO{
		ID:           u.ID,
		RestaurantID: u.RestaurantID,
		Name:         u.Name,
		Email:        u.Email,
		Role:         u.Role,
		Status:       u.Status,
	}
	response.Created(c, dto, "User created successfully")
}

func (h *Handler) DeleteUser(c *gin.Context) {
	restoID, _ := c.Get(auth.CtxRestaurantID)
	id := c.Param("id")

	if err := h.authRepo.Delete(c.Request.Context(), id, restoID.(string)); err != nil {
		response.InternalServerError(c, "DELETE_FAILED", err.Error())
		return
	}

	response.OK(c, gin.H{"deleted": true}, "User deleted successfully")
}
