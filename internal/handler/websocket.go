package handler

import (
	"errors"

	"github.com/JennerWork/chatbot/internal/service"
)

var (
	// ErrInvalidMessage 无效的消息
	ErrInvalidMessage = errors.New("无效的消息")
)

// WebSocketHandler WebSocket消息处理器
type WebSocketHandler struct {
	messageService service.MessageService
}

// NewWebSocketHandler 创建WebSocket消息处理器
func NewWebSocketHandler(messageService service.MessageService) *WebSocketHandler {
	return &WebSocketHandler{
		messageService: messageService,
	}
}

// HandleMessage 处理WebSocket消息
func (h *WebSocketHandler) HandleMessage(customerID uint, message []byte) ([]byte, error) {
	// TODO: 添加消息验证
	if len(message) == 0 {
		return nil, ErrInvalidMessage
	}

	// TODO: 添加消息处理前的钩子函数

	// 处理消息
	response, err := h.messageService.HandleMessage(customerID, message)
	if err != nil {
		return nil, err
	}

	// TODO: 添加消息处理后的钩子函数

	return response, nil
}
