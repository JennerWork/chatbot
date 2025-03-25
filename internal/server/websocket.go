package server

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/JennerWork/chatbot/internal/model"
	"github.com/gorilla/context"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// 1. 验证 Origin
		origin := r.Header.Get("Origin")
		// TODO: 从配置中获取允许的域名列表
		allowedOrigins := []string{"http://localhost:8080", "https://your-production-domain.com"}

		for _, allowed := range allowedOrigins {
			if origin == allowed {
				return true
			}
		}

		// 如果没有 Origin（比如来自非浏览器的客户端）且是本地请求，允许连接
		fmt.Println(r.Host)
		if origin == "" && (strings.Contains(r.Host, "localhost") || strings.Contains(r.Host, "127.0.0.1")) {
			return true
		}

		log.Printf("Rejected WebSocket connection from origin: %s", origin)
		return false
	},
}

// Client 表示一个WebSocket客户端连接
type Client struct {
	id           string             // 连接ID
	customerID   uint               // 客户ID
	conn         *websocket.Conn    // WebSocket连接
	send         chan []byte        // 发送消息的通道
	handlers     MessageHandlers    // 消息处理器
	lastActivity time.Time          // 最后活动时间
	manager      *ConnectionManager // 连接管理器
	session      *model.Session     // 关联的会话
	db           *gorm.DB           // 数据库连接
}

// MessageHandlers 定义消息处理器
type MessageHandlers interface {
	HandleMessage(customerID uint, message []byte) ([]byte, error)
}

// updateActivity 更新客户端活动时间
func (c *Client) updateActivity() {
	c.lastActivity = time.Now()
	// 更新会话最后活动时间
	c.session.LastActiveAt = c.lastActivity
	c.db.Save(c.session)
}

// HandleWebSocket 处理WebSocket连接请求
func (cm *ConnectionManager) HandleWebSocket(w http.ResponseWriter, r *http.Request, handlers MessageHandlers) {
	// 从 context 获取认证信息
	customerID := getCustomerIDFromRequest(r)
	if customerID == 0 {
		log.Printf("Unauthorized WebSocket connection attempt")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 升级连接
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}

	// 创建新会话，初始状态为initiated
	session := &model.Session{
		CustomerID:   customerID,
		Status:       string(model.SessionStatusInitiated),
		LastActiveAt: time.Now(),
	}

	if err := cm.db.Create(session).Error; err != nil {
		log.Printf("Failed to create session: %v", err)
		conn.Close()
		return
	}

	log.Printf("New session initiated for customer %d", customerID)

	client := &Client{
		conn:         conn,
		send:         make(chan []byte, 256),
		handlers:     handlers,
		customerID:   customerID,
		lastActivity: time.Now(),
		manager:      cm,
		session:      session,
		db:           cm.db,
	}

	// 注册客户端（这里会将状态更新为active）
	cm.Register(client)

	// 启动读写goroutine
	go client.writePump()
	go client.readPump()
}

// writePump 将消息发送到WebSocket连接
func (c *Client) writePump() {
	defer func() {
		c.conn.Close()
		c.manager.Unregister(c)
	}()

	for {
		select {
		case msg, ok := <-c.send:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, nil)
				return
			}
			c.conn.SetWriteDeadline(time.Now().Add(time.Second * 10))
			if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}
			c.updateActivity()
		}
	}
}

// readPump 从WebSocket连接读取消息
func (c *Client) readPump() {
	defer func() {
		c.manager.Unregister(c)
		close(c.send)
		c.conn.Close()
	}()

	// 设置读取超时
	c.conn.SetReadDeadline(time.Now().Add(time.Second * 60))
	c.conn.SetPongHandler(func(string) error {
		c.updateActivity()
		c.conn.SetReadDeadline(time.Now().Add(time.Second * 60))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket read error for customer %d: %v", c.customerID, err)
			}
			break
		}

		c.updateActivity()
		c.conn.SetReadDeadline(time.Now().Add(time.Second * 60))

		// 处理接收到的消息并发送回复
		response, err := c.handlers.HandleMessage(c.customerID, message)
		if err != nil {
			log.Printf("Error handling message for customer %d: %v", c.customerID, err)
			continue
		}

		// 如果有回复消息，发送给客户端
		if response != nil {
			c.send <- response
		}
	}
}

// getCustomerIDFromRequest 从请求中获取客户ID
func getCustomerIDFromRequest(r *http.Request) uint {
	// 从gin的Context中获取customerID
	customerID := context.Get(r, "customer_id")
	if customerID != nil {
		if customerID, ok := customerID.(uint); ok {
			return customerID
		}
	}
	return 0
}
