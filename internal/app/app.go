package app

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/JennerWork/chatbot/internal/config"
	"github.com/JennerWork/chatbot/internal/handler"
	"github.com/JennerWork/chatbot/internal/server"
	"github.com/JennerWork/chatbot/internal/service"
	"github.com/JennerWork/chatbot/pkg/db"
)

// Run 启动应用程序
func Run(configPath string) error {
	log.Printf("Starting chatbot server...")
	startTime := time.Now()

	// 加载配置
	log.Printf("Loading configuration from %s", configPath)
	if err := config.LoadConfig(configPath); err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}
	log.Printf("Configuration loaded successfully")

	// 初始化数据库连接
	log.Printf("Initializing database connection to %s:%d...",
		config.GlobalConfig.Database.Host,
		config.GlobalConfig.Database.Port)
	if err := db.Init(&config.GlobalConfig.Database); err != nil {
		return fmt.Errorf("failed to initialize database: %v", err)
	}
	log.Printf("Database connection established successfully")

	// 获取数据库连接
	dbConn := db.GetDB()

	// 创建消息服务
	log.Printf("Initializing message service...")
	msgService := service.NewMessageService(dbConn)
	log.Printf("Message service initialized")

	// 创建连接管理器
	log.Printf("Creating connection manager...")
	cm := server.NewConnectionManager(dbConn)
	log.Printf("Connection manager created")

	// 创建消息处理器
	log.Printf("Setting up WebSocket handlers...")
	handlers := handler.NewWebSocketHandler(msgService)
	log.Printf("WebSocket handlers initialized")

	// 创建HTTP服务器
	log.Printf("Creating HTTP server...")
	srv := server.NewServer()
	srv.SetupRoutes(dbConn, handlers, cm)
	log.Printf("HTTP server created and routes configured")

	// 启动定期清理不活跃连接的goroutine
	log.Printf("Starting inactive connection cleanup routine...")
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			log.Printf("Running cleanup of inactive connections...")
			cm.CleanInactiveConnections(30 * time.Minute)
		}
	}()

	// 启动HTTP服务器
	log.Printf("Starting HTTP server on port %d...", config.GlobalConfig.App.Port)
	go func() {
		if err := srv.Start(config.GlobalConfig.App.Port); err != nil {
			log.Printf("Server error: %v", err)
		}
	}()

	// 计算启动时间
	startupDuration := time.Since(startTime)
	log.Printf("Server startup completed in %.2f seconds", startupDuration.Seconds())
	log.Printf("Server is ready to accept connections at http://localhost:%d", config.GlobalConfig.App.Port)

	// 等待中断信号优雅地关闭服务器
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Received shutdown signal, initiating graceful shutdown...")

	// 设置关闭超时时间
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 关闭HTTP服务器
	log.Println("Shutting down HTTP server...")
	if err := srv.Stop(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server shutdown completed successfully")
	return nil
}
