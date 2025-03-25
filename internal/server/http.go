package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Server struct {
	httpServer *http.Server
	router     *gin.Engine
}

// NewServer 创建新的服务器实例
func NewServer() *Server {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	return &Server{
		router: router,
	}
}

// Start 启动HTTP服务器
func (s *Server) Start(port int) error {
	s.httpServer = &http.Server{
		Addr:           fmt.Sprintf(":%d", port),
		Handler:        s.router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	return s.httpServer.ListenAndServe()
}

// Stop 优雅地停止HTTP服务器
func (s *Server) Stop(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

// Router 返回gin路由实例
func (s *Server) Router() *gin.Engine {
	return s.router
}
