package handler

import (
	"net/http"
	"strings"

	"github.com/JennerWork/chatbot/internal/service"
	"github.com/gin-gonic/gin"
)

// LoginRequest login request parameters
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// TokenResponse token response
type TokenResponse struct {
	Token string `json:"token"`
}

// Login login handler
func Login(authService service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Code:    400,
				Message: "Invalid request parameters",
				Error:   err.Error(),
			})
			return
		}

		token, err := authService.Login(req.Email, req.Password)
		if err != nil {
			status := http.StatusInternalServerError
			message := "Login failed"

			if err == service.ErrInvalidCredentials {
				status = http.StatusUnauthorized
				message = "Incorrect email or password"
			}

			c.JSON(status, ErrorResponse{
				Code:    status,
				Message: message,
				Error:   err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, TokenResponse{
			Token: token,
		})
	}
}

// RefreshToken token refresh handler
func RefreshToken(authService service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, ErrorResponse{
				Code:    401,
				Message: "Authentication token not provided",
			})
			return
		}

		// 解析token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, ErrorResponse{
				Code:    401,
				Message: "Invalid token format",
			})
			return
		}

		// 刷新token
		newToken, err := authService.RefreshToken(parts[1])
		if err != nil {
			status := http.StatusInternalServerError
			message := "Failed to refresh token"

			switch err {
			case service.ErrTokenExpired:
				status = http.StatusUnauthorized
				message = "Token expired"
			case service.ErrInvalidToken:
				status = http.StatusUnauthorized
				message = "Invalid token"
			}

			c.JSON(status, ErrorResponse{
				Code:    status,
				Message: message,
				Error:   err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, TokenResponse{
			Token: newToken,
		})
	}
}
