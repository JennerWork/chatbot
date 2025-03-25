package model

import (
	"time"

	"gorm.io/gorm"
)

type Session struct {
	gorm.Model
	CustomerID   uint      `json:"customer_id"`
	Customer     Customer  `json:"customer"`
	Status       string    `json:"status"`
	LastActiveAt time.Time `json:"last_active_at"`
	Messages     []Message `json:"messages"`
}

// SessionStatus 定义会话状态
type SessionStatus string

// 会话状态转换规则：
// 1. initiated -> active：
//    - 当WebSocket连接成功建立时
//    - 当客户端首次连接时
//
// 2. active -> inactive：
//    - 当连接超过30分钟无活动时
//    - 当连接发生非正常断开时
//
// 3. active/inactive -> cancelled：
//    - 当用户主动关闭连接时
//    - 当同一用户建立新连接，旧连接被关闭时

const (
	// SessionStatusInitiated 会话已创建但未开始
	SessionStatusInitiated SessionStatus = "initiated"
	// SessionStatusActive 会话活跃中
	SessionStatusActive SessionStatus = "active"
	// SessionStatusInactive 会话因超时或无活动而结束
	SessionStatusInactive SessionStatus = "inactive"
	// SessionStatusCancelled 会话被用户主动关闭
	SessionStatusCancelled SessionStatus = "cancelled"
)
