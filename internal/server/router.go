package server

import (
	"net/http"
	"time"

	"github.com/JennerWork/chatbot/internal/handler"
	"github.com/JennerWork/chatbot/internal/middleware"
	"github.com/JennerWork/chatbot/internal/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SetupRoutes 配置所有路由
func (s *Server) SetupRoutes(db *gorm.DB, handlers MessageHandlers, cm *ConnectionManager) {
	// 创建服务实例
	messageQueryService := service.NewMessageQueryService(db)
	messageHandler := handler.NewMessageHandler(messageQueryService)

	// 创建认证服务
	jwtConfig := service.JWTConfig{
		SecretKey:     "your-secret-key",  // TODO: 从配置文件读取
		TokenExpiry:   time.Hour * 24,     // token有效期24小时
		RefreshExpiry: time.Hour * 24 * 7, // 刷新token有效期7天
	}
	authService := service.NewAuthService(db, jwtConfig)

	// 创建认证中间件
	authMiddleware := middleware.AuthMiddleware(authService)

	// WebSocket路由（需要认证）
	s.router.GET("/ws", authMiddleware, func(c *gin.Context) {
		cm.HandleWebSocket(c.Writer, c.Request, handlers)
	})

	// 健康检查（无需认证）
	s.router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":      "ok",
			"connections": cm.GetActiveConnections(),
		})
	})

	// API路由组
	api := s.router.Group("/api")
	{
		// 认证相关路由（无需认证）
		auth := api.Group("/auth")
		{
			auth.POST("/login", handler.Login(authService))
			auth.POST("/refresh", handler.RefreshToken(authService))
		}

		// 需要认证的路由组
		authenticated := api.Group("")
		authenticated.Use(authMiddleware)
		{
			// 消息相关路由
			messages := authenticated.Group("/message")
			{
				messages.GET("/list", messageHandler.GetMessageHistory)
			}
		}
	}
}
