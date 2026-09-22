package auth

import (
	"strings"

	"github.com/gin-gonic/gin"
	"qr-store/backend/pkg/response"
)

const (
	CtxUserID       = "user_id"
	CtxRestaurantID = "restaurant_id"
	CtxRole         = "role"
)

func AuthMiddleware(service Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "UNAUTHORIZED", "Missing authorization header")
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Unauthorized(c, "INVALID_TOKEN_FORMAT", "Invalid authorization header format")
			c.Abort()
			return
		}

		tokenString := parts[1]
		claims, err := service.ValidateToken(tokenString)
		if err != nil {
			response.Unauthorized(c, "INVALID_TOKEN", "Token is invalid or expired")
			c.Abort()
			return
		}

		c.Set(CtxUserID, claims.UserID)
		c.Set(CtxRestaurantID, claims.RestaurantID)
		c.Set(CtxRole, claims.Role)
		c.Next()
	}
}

func RequireRoles(allowedRoles ...Role) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleVal, exists := c.Get(CtxRole)
		if !exists {
			response.Forbidden(c, "FORBIDDEN", "User role not found")
			c.Abort()
			return
		}

		userRole, ok := roleVal.(Role)
		if !ok {
			response.Forbidden(c, "FORBIDDEN", "Invalid user role")
			c.Abort()
			return
		}

		for _, role := range allowedRoles {
			if userRole == role {
				c.Next()
				return
			}
		}

		response.Forbidden(c, "INSUFFICIENT_PERMISSIONS", "You do not have permission to access this resource")
		c.Abort()
	}
}
