package middleware

import (
	"net/http"
	"strings"

	"github.com/JennerWork/chatbot/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/context"
)

// AuthMiddleware 创建认证中间件
func AuthMiddleware(authService service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "未提供认证token",
			})
			c.Abort()
			return
		}

		// 解析token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    401,
				"message": "无效的token格式",
			})
			c.Abort()
			return
		}

		// 验证token
		claims, err := authService.ValidateToken(parts[1])
		if err != nil {
			var status int
			var message string

			switch err {
			case service.ErrTokenExpired:
				status = http.StatusUnauthorized
				message = "token已过期"
			case service.ErrInvalidToken:
				status = http.StatusUnauthorized
				message = "无效的token"
			default:
				status = http.StatusInternalServerError
				message = "token验证失败"
			}

			c.JSON(status, gin.H{
				"code":    status,
				"message": message,
				"error":   err.Error(),
			})
			c.Abort()
			return
		}

		// 将用户信息存储到上下文中
		c.Set("customer_id", claims.CustomerID)
		context.Set(c.Request, "customer_id", claims.CustomerID)
		c.Set("email", claims.Email)
		context.Set(c.Request, "email", claims.Email)
		c.Next()
	}
}

// GetCustomerID 从上下文中获取客户ID
func GetCustomerID(c *gin.Context) uint {
	if id, exists := c.Get("customer_id"); exists {
		if customerID, ok := id.(uint); ok {
			return customerID
		}
	}
	return 0
}

// GetCustomerEmail 从上下文中获取客户邮箱
func GetCustomerEmail(c *gin.Context) string {
	if email, exists := c.Get("email"); exists {
		if customerEmail, ok := email.(string); ok {
			return customerEmail
		}
	}
	return ""
}
