package service

import (
	"encoding/json"
	"time"

	"github.com/JennerWork/chatbot/internal/model"
	"gorm.io/gorm"
)

// MessageRequest 定义客户端发送的消息格式
type MessageRequest struct {
	Type    string          `json:"type"`            // 消息类型：text, image, etc.
	Content json.RawMessage `json:"content"`         // 消息内容，根据type解析
	Extra   json.RawMessage `json:"extra,omitempty"` // 额外参数
}

// MessageResponse 定义返回给客户端的消息格式
type MessageResponse struct {
	Type      string          `json:"type"`            // 消息类型
	Content   json.RawMessage `json:"content"`         // 消息内容
	Timestamp time.Time       `json:"timestamp"`       // 消息时间戳
	Extra     json.RawMessage `json:"extra,omitempty"` // 额外参数
}

// MessageService 定义消息处理服务的接口
type MessageService interface {
	// HandleMessage 处理接收到的消息
	HandleMessage(customerID uint, message []byte) ([]byte, error)
}

// messageService 实现 MessageService 接口
type messageService struct {
	db          *gorm.DB
	chatService ChatService
}

// NewMessageService 创建新的消息服务实例
func NewMessageService(db *gorm.DB) MessageService {
	return &messageService{
		db:          db,
		chatService: NewChatService(db),
	}
}

// HandleMessage 处理消息的具体实现
func (s *messageService) HandleMessage(customerID uint, message []byte) ([]byte, error) {
	// 1. 解析接收到的消息
	var request MessageRequest
	if err := json.Unmarshal(message, &request); err != nil {
		return nil, err
	}

	// 2. 获取或创建当前会话
	var session model.Session
	if err := s.db.Where("customer_id = ? AND status = ?",
		customerID, model.SessionStatusActive).
		First(&session).Error; err != nil {
		// 如果没有活跃会话，创建新会话
		session = model.Session{
			CustomerID:   customerID,
			Status:       string(model.SessionStatusActive),
			LastActiveAt: time.Now(),
		}
		if err := s.db.Create(&session).Error; err != nil {
			return nil, err
		}
	}

	// 3. 保存客户发送的消息
	dbMessage := &model.Message{
		CustomerID: customerID,
		SessionID:  session.ID,
		Content:    string(request.Content),
		Sender:     model.SenderCustomer,
		Seq:        s.getNextMessageSeq(session.ID),
	}
	if err := s.db.Create(dbMessage).Error; err != nil {
		return nil, err
	}

	// 4. 根据消息类型处理消息
	response, err := s.processMessage(customerID, session.ID, &request)
	if err != nil {
		return nil, err
	}

	// 5. 保存机器人的回复
	botMessage := &model.Message{
		CustomerID: customerID,
		SessionID:  session.ID,
		Content:    string(response.Content),
		Sender:     model.SenderBot,
		Seq:        s.getNextMessageSeq(session.ID),
	}
	if err := s.db.Create(botMessage).Error; err != nil {
		return nil, err
	}

	// 6. 更新会话最后活动时间
	session.LastActiveAt = time.Now()
	if err := s.db.Save(&session).Error; err != nil {
		return nil, err
	}

	// 7. 将响应序列化为JSON
	return json.Marshal(response)
}

// getNextMessageSeq 获取下一个消息序号
func (s *messageService) getNextMessageSeq(sessionID uint) uint {
	var maxSeq struct {
		MaxSeq uint
	}
	s.db.Model(&model.Message{}).
		Select("COALESCE(MAX(seq), 0) as max_seq").
		Where("session_id = ?", sessionID).
		Scan(&maxSeq)
	return maxSeq.MaxSeq + 1
}

// processMessage 根据消息类型处理消息
func (s *messageService) processMessage(customerID uint, sessionID uint, request *MessageRequest) (*MessageResponse, error) {
	switch request.Type {
	case "text":
		return s.handleTextMessage(customerID, sessionID, request)
	default:
		return s.handleUnknownMessage(customerID, request)
	}
}

// handleTextMessage 处理文本消息
func (s *messageService) handleTextMessage(customerID uint, sessionID uint, request *MessageRequest) (*MessageResponse, error) {
	// 1. 解析文本消息内容
	var textMsg TextMessage
	if err := json.Unmarshal(request.Content, &textMsg); err != nil {
		return nil, err
	}

	// 2. 使用聊天服务处理文本消息
	reply, err := s.chatService.ProcessText(customerID, sessionID, textMsg.Text)
	if err != nil {
		return nil, err
	}

	// 3. 构造响应
	content, err := json.Marshal(TextMessage{Text: reply})
	if err != nil {
		return nil, err
	}

	return &MessageResponse{
		Type:      "text",
		Content:   content,
		Timestamp: time.Now(),
	}, nil
}

// handleUnknownMessage 处理未知类型的消息
func (s *messageService) handleUnknownMessage(customerID uint, request *MessageRequest) (*MessageResponse, error) {
	content, _ := json.Marshal(TextMessage{Text: "抱歉，暂不支持该类型的消息"})

	return &MessageResponse{
		Type:      "text",
		Content:   content,
		Timestamp: time.Now(),
	}, nil
}
