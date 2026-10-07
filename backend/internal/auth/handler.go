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
		authGroup.PUT("/password", AuthMiddleware(h.service), h.ChangePassword)
		authGroup.POST("/change-password", AuthMiddleware(h.service), h.ChangePassword)
		authGroup.PUT("/account", AuthMiddleware(h.service), h.UpdateAccount)
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

func (h *Handler) ChangePassword(c *gin.Context) {
	userID, _ := c.Get(CtxUserID)
	idStr, ok := userID.(string)
	if !ok || idStr == "" {
		response.Unauthorized(c, "UNAUTHORIZED", "User ID not found in session")
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_REQUEST", err.Error())
		return
	}

	if err := h.service.ChangePassword(c.Request.Context(), idStr, req.CurrentPassword, req.NewPassword); err != nil {
		if errors.Is(err, ErrInvalidCurrentPassword) {
			response.BadRequest(c, "INVALID_CURRENT_PASSWORD", "Kata sandi saat ini salah")
			return
		}
		response.InternalServerError(c, "CHANGE_PASSWORD_FAILED", "Gagal memperbarui kata sandi")
		return
	}

	response.OK(c, gin.H{"updated": true}, "Kata sandi berhasil diperbarui")
}

func (h *Handler) UpdateAccount(c *gin.Context) {
	userID, _ := c.Get(CtxUserID)
	idStr, ok := userID.(string)
	if !ok || idStr == "" {
		response.Unauthorized(c, "UNAUTHORIZED", "User ID not found in session")
		return
	}

	var req UpdateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_REQUEST", err.Error())
		return
	}

	resp, err := h.service.UpdateAccount(c.Request.Context(), idStr, req)
	if err != nil {
		if errors.Is(err, ErrInvalidCurrentPassword) {
			response.BadRequest(c, "INVALID_CURRENT_PASSWORD", "Kata sandi saat ini salah")
			return
		}
		if errors.Is(err, ErrEmailAlreadyExists) {
			response.BadRequest(c, "EMAIL_EXISTS", "Alamat email sudah digunakan oleh akun lain")
			return
		}
		if errors.Is(err, ErrIDAlreadyExists) {
			response.BadRequest(c, "ID_EXISTS", "ID pengguna sudah digunakan")
			return
		}
		response.InternalServerError(c, "UPDATE_FAILED", err.Error())
		return
	}

	response.OK(c, resp, "Akun berhasil diperbarui")
}

