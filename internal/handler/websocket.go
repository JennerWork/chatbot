package handler

import (
	"errors"

	"github.com/JennerWork/chatbot/internal/service"
)

var (
	// ErrInvalidMessage invalid message
	ErrInvalidMessage = errors.New("无效的消息")
)

// WebSocketHandler WebSocket message handler
type WebSocketHandler struct {
	messageService service.MessageService
}

// NewWebSocketHandler create WebSocket message handler
func NewWebSocketHandler(messageService service.MessageService) *WebSocketHandler {
	return &WebSocketHandler{
		messageService: messageService,
	}
}

// HandleMessage handle WebSocket message
func (h *WebSocketHandler) HandleMessage(customerID uint, message []byte) ([]byte, error) {
	// TODO: Add message validation
	if len(message) == 0 {
		return nil, ErrInvalidMessage
	}

	// TODO: Add pre-message processing hooks

	// 处理消息
	response, err := h.messageService.HandleMessage(customerID, message)
	if err != nil {
		return nil, err
	}

	// TODO: Add post-message processing hooks

	return response, nil
}
