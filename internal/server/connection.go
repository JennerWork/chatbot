package server

import (
	"log"
	"sync"
	"time"

	"github.com/JennerWork/chatbot/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ConnectionManager 管理所有的WebSocket连接和会话
type ConnectionManager struct {
	connections map[string]*Client // 连接ID -> 客户端连接
	sessions    map[uint]*Client   // 用户ID -> 客户端连接
	mu          sync.RWMutex
	db          *gorm.DB // 数据库连接
}

// NewConnectionManager 创建新的连接管理器
func NewConnectionManager(db *gorm.DB) *ConnectionManager {
	cm := &ConnectionManager{
		connections: make(map[string]*Client),
		sessions:    make(map[uint]*Client),
		db:          db,
	}

	// 启动定期清理协程
	go cm.startCleanupLoop()
	return cm
}

// startCleanupLoop 启动定期清理循环
func (cm *ConnectionManager) startCleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		cm.CleanInactiveConnections(30 * time.Minute)
	}
}

// Register 注册新的客户端连接
func (cm *ConnectionManager) Register(client *Client) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// 生成连接ID
	connectionID := uuid.New().String()
	client.id = connectionID
	cm.connections[connectionID] = client

	// 如果有客户ID，建立客户会话映射
	if client.customerID > 0 {
		// 如果客户已有连接，关闭旧连接
		if oldClient, exists := cm.sessions[client.customerID]; exists {
			// 更新旧会话状态为已关闭
			if oldClient.session != nil {
				oldClient.session.Status = string(model.SessionStatusCancelled)
				cm.db.Save(oldClient.session)
				log.Printf("Session cancelled for customer %d (replaced by new connection)", oldClient.customerID)
			}
			close(oldClient.send)
			delete(cm.connections, oldClient.id)
		}
		cm.sessions[client.customerID] = client
	}

	// 更新新会话状态为活跃
	if client.session != nil {
		client.session.Status = string(model.SessionStatusActive)
		cm.db.Save(client.session)
		log.Printf("Session activated for customer %d", client.customerID)
	}
}

// Unregister 注销客户端连接
func (cm *ConnectionManager) Unregister(client *Client) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// 更新会话状态
	if client.session != nil {
		client.session.Status = string(model.SessionStatusCancelled)
		cm.db.Save(client.session)
		log.Printf("Session cancelled for customer %d (unregistered)", client.customerID)
	}

	if client.customerID > 0 {
		delete(cm.sessions, client.customerID)
	}
	delete(cm.connections, client.id)
}

// GetClient 根据连接ID获取客户端
func (cm *ConnectionManager) GetClient(connectionID string) (*Client, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	client, exists := cm.connections[connectionID]
	return client, exists
}

// GetClientByCustomerID 根据客户ID获取客户端
func (cm *ConnectionManager) GetClientByCustomerID(customerID uint) (*Client, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	client, exists := cm.sessions[customerID]
	return client, exists
}

// GetActiveConnections 获取活跃连接数
func (cm *ConnectionManager) GetActiveConnections() int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return len(cm.connections)
}

// CleanInactiveConnections 清理不活跃的连接
func (cm *ConnectionManager) CleanInactiveConnections(inactiveTimeout time.Duration) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	now := time.Now()
	for _, client := range cm.connections {
		if now.Sub(client.lastActivity) > inactiveTimeout {
			log.Printf("Closing inactive connection for customer %d, last activity: %v",
				client.customerID, client.lastActivity)

			// 更新会话状态为不活跃
			if client.session != nil {
				client.session.Status = string(model.SessionStatusInactive)
				cm.db.Save(client.session)
				log.Printf("Session marked as inactive for customer %d (timeout after %v)",
					client.customerID, inactiveTimeout)
			}

			// 关闭连接
			client.conn.Close()
			delete(cm.connections, client.id)
			if client.customerID > 0 {
				delete(cm.sessions, client.customerID)
			}
		}
	}
}
