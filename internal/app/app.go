package app

import (
	"context"
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
	// 加载配置
	if err := config.LoadConfig(configPath); err != nil {
		return err
	}

	// 初始化数据库连接
	if err := db.Init(&config.GlobalConfig.Database); err != nil {
		return err
	}

	// 获取数据库连接
	dbConn := db.GetDB()

	// 创建消息服务
	msgService := service.NewMessageService(dbConn)

	// 创建连接管理器
	cm := server.NewConnectionManager(dbConn)

	// 创建消息处理器
	handlers := handler.NewWebSocketHandler(msgService)

	// 创建HTTP服务器
	srv := server.NewServer()
	srv.SetupRoutes(dbConn, handlers, cm)

	// 启动定期清理不活跃连接的goroutine
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			cm.CleanInactiveConnections(30 * time.Minute)
		}
	}()

	// 启动HTTP服务器
	go func() {
		if err := srv.Start(config.GlobalConfig.App.Port); err != nil {
			log.Printf("Server error: %v", err)
		}
	}()

	// 等待中断信号优雅地关闭服务器
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// 设置关闭超时时间
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 关闭HTTP服务器
	if err := srv.Stop(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
	return nil
}
