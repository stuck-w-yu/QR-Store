package auth

import (
	"errors"
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
	authGroup := r.Group("/auth")
	{
		authGroup.POST("/login", h.Login)
		authGroup.GET("/me", AuthMiddleware(h.service), h.Me)
		authGroup.POST("/logout", AuthMiddleware(h.service), h.Logout)
	}
}

func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_REQUEST", err.Error())
		return
	}

	authResp, err := h.service.Login(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			response.Unauthorized(c, "INVALID_CREDENTIALS", "Invalid email or password")
			return
		}
		response.InternalServerError(c, "LOGIN_FAILED", "An error occurred while logging in")
		return
	}

	response.OK(c, authResp, "Login successful")
}

func (h *Handler) Me(c *gin.Context) {
	userID, _ := c.Get(CtxUserID)
	idStr, ok := userID.(string)
	if !ok || idStr == "" {
		response.Unauthorized(c, "UNAUTHORIZED", "User ID not found in session")
		return
	}

	user, err := h.service.GetUserByID(c.Request.Context(), idStr)
	if err != nil {
		response.NotFound(c, "USER_NOT_FOUND", "User profile not found")
		return
	}

	response.OK(c, user, "User profile retrieved")
}

func (h *Handler) Logout(c *gin.Context) {
	response.JSON(c, http.StatusOK, gin.H{"logged_out": true}, "Successfully logged out")
}
