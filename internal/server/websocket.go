package server

import (
	"log"
	"net/http"
	"time"

	"github.com/JennerWork/chatbot/internal/model"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // 在生产环境中应该根据实际需求设置允许的来源
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

// HandleWebSocket 处理WebSocket连接请求
func (cm *ConnectionManager) HandleWebSocket(w http.ResponseWriter, r *http.Request, handlers MessageHandlers) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}

	// 从请求中获取客户ID
	customerID := getCustomerIDFromRequest(r)

	// 创建或更新会话
	session := &model.Session{
		CustomerID:   customerID,
		Status:       string(model.SessionStatusActive),
		LastActiveAt: time.Now(),
	}

	if err := cm.db.Create(session).Error; err != nil {
		log.Printf("Failed to create session: %v", err)
		conn.Close()
		return
	}

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

	// 注册客户端
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

	for message := range c.send {
		w, err := c.conn.NextWriter(websocket.TextMessage)
		if err != nil {
			return
		}
		w.Write(message)

		if err := w.Close(); err != nil {
			return
		}
		c.lastActivity = time.Now()

		// 更新会话最后活动时间
		c.session.LastActiveAt = c.lastActivity
		c.db.Save(c.session)
	}

	c.conn.WriteMessage(websocket.CloseMessage, []byte{})
}

// readPump 从WebSocket连接读取消息
func (c *Client) readPump() {
	defer func() {
		c.manager.Unregister(c)
		close(c.send)
		c.conn.Close()
	}()

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		c.lastActivity = time.Now()

		// 更新会话最后活动时间
		c.session.LastActiveAt = c.lastActivity
		c.db.Save(c.session)

		// 处理接收到的消息并发送回复
		response, err := c.handlers.HandleMessage(c.customerID, message)
		if err != nil {
			log.Printf("Error handling message: %v", err)
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
	// TODO: 实现从请求中获取客户ID的逻辑
	// 可以从cookie、header或query参数中获取
	return 0
}
