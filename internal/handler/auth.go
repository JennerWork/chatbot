package handler

import (
	"net/http"
	"strings"

	"github.com/JennerWork/chatbot/internal/service"
	"github.com/gin-gonic/gin"
)

// LoginRequest 登录请求参数
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// TokenResponse token响应
type TokenResponse struct {
	Token string `json:"token"`
}

// Login 登录处理器
func Login(authService service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req LoginRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Code:    400,
				Message: "无效的请求参数",
				Error:   err.Error(),
			})
			return
		}

		token, err := authService.Login(req.Email, req.Password)
		if err != nil {
			status := http.StatusInternalServerError
			message := "登录失败"

			if err == service.ErrInvalidCredentials {
				status = http.StatusUnauthorized
				message = "邮箱或密码错误"
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

// RefreshToken token刷新处理器
func RefreshToken(authService service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, ErrorResponse{
				Code:    401,
				Message: "未提供认证token",
			})
			return
		}

		// 解析token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, ErrorResponse{
				Code:    401,
				Message: "无效的token格式",
			})
			return
		}

		// 刷新token
		newToken, err := authService.RefreshToken(parts[1])
		if err != nil {
			status := http.StatusInternalServerError
			message := "刷新token失败"

			switch err {
			case service.ErrTokenExpired:
				status = http.StatusUnauthorized
				message = "token已过期"
			case service.ErrInvalidToken:
				status = http.StatusUnauthorized
				message = "无效的token"
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
