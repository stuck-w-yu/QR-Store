package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

type APIErrorResponse struct {
	Success bool     `json:"success"`
	Error   APIError `json:"error"`
}

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func JSON(c *gin.Context, httpStatus int, data interface{}, message string) {
	c.JSON(httpStatus, APIResponse{
		Success: true,
		Data:    data,
		Message: message,
	})
}

func OK(c *gin.Context, data interface{}, message string) {
	if message == "" {
		message = "Success"
	}
	JSON(c, http.StatusOK, data, message)
}

func Created(c *gin.Context, data interface{}, message string) {
	if message == "" {
		message = "Created successfully"
	}
	JSON(c, http.StatusCreated, data, message)
}

func Error(c *gin.Context, httpStatus int, code string, message string) {
	c.JSON(httpStatus, APIErrorResponse{
		Success: false,
		Error: APIError{
			Code:    code,
			Message: message,
		},
	})
}

func BadRequest(c *gin.Context, code string, message string) {
	if code == "" {
		code = "BAD_REQUEST"
	}
	Error(c, http.StatusBadRequest, code, message)
}

func Unauthorized(c *gin.Context, code string, message string) {
	if code == "" {
		code = "UNAUTHORIZED"
	}
	Error(c, http.StatusUnauthorized, code, message)
}

func Forbidden(c *gin.Context, code string, message string) {
	if code == "" {
		code = "FORBIDDEN"
	}
	Error(c, http.StatusForbidden, code, message)
}

func NotFound(c *gin.Context, code string, message string) {
	if code == "" {
		code = "NOT_FOUND"
	}
	Error(c, http.StatusNotFound, code, message)
}

func Conflict(c *gin.Context, code string, message string) {
	if code == "" {
		code = "CONFLICT"
	}
	Error(c, http.StatusConflict, code, message)
}

func InternalServerError(c *gin.Context, code string, message string) {
	if code == "" {
		code = "INTERNAL_SERVER_ERROR"
	}
	Error(c, http.StatusInternalServerError, code, message)
}
