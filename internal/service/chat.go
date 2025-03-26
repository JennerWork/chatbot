package service

import (
	"fmt"
	"strings"

	"github.com/JennerWork/chatbot/internal/model"
	"gorm.io/gorm"
)

// TextMessage 文本消息的具体内容
type TextMessage struct {
	Text string `json:"text"`
}

// ChatService 处理聊天相关的业务逻辑
type ChatService interface {
	// ProcessText 处理文本消息并返回回复
	ProcessText(customerID uint, sessionID uint, text string) (string, error)
}

type chatService struct {
	db *gorm.DB
}

// NewChatService 创建聊天服务实例
func NewChatService(db *gorm.DB) ChatService {
	return &chatService{
		db: db,
	}
}

// ProcessText 处理文本消息
func (s *chatService) ProcessText(customerID uint, sessionID uint, text string) (string, error) {
	// 1. 获取当前会话的上下文
	var session model.Session
	if err := s.db.First(&session, sessionID).Error; err != nil {
		return "", fmt.Errorf("获取会话失败: %v", err)
	}

	// 2. 检查是否有未完成的反馈
	var feedback model.Feedback
	result := s.db.Where("customer_id = ? AND status != ?",
		customerID, model.FeedbackStatusCompleted).
		Last(&feedback)
	hasPendingFeedback := result.Error == nil

	if hasPendingFeedback {
		// 处理反馈流程中的回复
		return s.handleFeedbackResponse(text, &feedback)
	}

	// 3. 检查是否是反馈触发词
	if s.isFeedbackTrigger(text) {
		return s.initiateFeedback(customerID, sessionID)
	}

	// 4. 获取会话历史消息，用于上下文理解
	var recentMessages []model.Message
	if err := s.db.Where("session_id = ?", sessionID).
		Order("created_at desc").
		Find(&recentMessages).Error; err != nil {
		return "", fmt.Errorf("获取历史消息失败: %v", err)
	}

	// 5. 生成普通回复
	return s.generateBasicReply(text, recentMessages), nil
}

// isFeedbackTrigger 检查是否是反馈触发词
func (s *chatService) isFeedbackTrigger(text string) bool {
	triggers := []string{"feedback", "review", "评价", "反馈", "评论"}
	text = strings.ToLower(text)
	for _, trigger := range triggers {
		if strings.Contains(text, trigger) {
			return true
		}
	}
	return false
}

// initiateFeedback 初始化反馈流程
func (s *chatService) initiateFeedback(customerID uint, sessionID uint) (string, error) {
	// 创建新的反馈记录
	feedback := model.Feedback{
		CustomerID: customerID,
		SessionID:  sessionID,
		Status:     model.FeedbackStatusInitiated,
	}

	if err := s.db.Create(&feedback).Error; err != nil {
		return "", fmt.Errorf("创建反馈记录失败: %v", err)
	}

	return "请对本次服务进行评分（1-5分），5分表示非常满意，1分表示非常不满意", nil
}

// handleFeedbackResponse 处理反馈流程中的回复
func (s *chatService) handleFeedbackResponse(text string, feedback *model.Feedback) (string, error) {
	switch feedback.Status {
	case model.FeedbackStatusInitiated:
		// 处理评分
		rating, err := s.parseRating(text)
		if err != nil {
			return "请输入1-5之间的数字进行评分", nil
		}

		feedback.Rating = rating
		feedback.Status = model.FeedbackStatusRatingProvided
		if err := s.db.Save(feedback).Error; err != nil {
			return "", fmt.Errorf("保存评分失败: %v", err)
		}

		return "感谢您的评分！请问您对我们的服务有什么建议或意见吗？", nil

	case model.FeedbackStatusRatingProvided:
		// 处理评论
		sentimentService := NewSentimentAnalysisService()
		sentiment := sentimentService.AnalyzeSentiment(text, int(feedback.Rating))
		feedback.Comment = text
		feedback.Sentiment = sentiment
		feedback.Status = model.FeedbackStatusCompleted
		if err := s.db.Save(feedback).Error; err != nil {
			return "", fmt.Errorf("保存评论失败: %v", err)
		}

		return "非常感谢您的反馈！我们会继续努力提供更好的服务。", nil

	default:
		return s.generateBasicReply(text, nil), nil
	}
}

// parseRating 解析评分
func (s *chatService) parseRating(text string) (model.FeedbackRating, error) {
	text = strings.TrimSpace(text)
	switch text {
	case "1":
		return model.FeedbackRating1, nil
	case "2":
		return model.FeedbackRating2, nil
	case "3":
		return model.FeedbackRating3, nil
	case "4":
		return model.FeedbackRating4, nil
	case "5":
		return model.FeedbackRating5, nil
	default:
		return 0, fmt.Errorf("无效的评分")
	}
}

// generateBasicReply 生成基础回复
func (s *chatService) generateBasicReply(text string, context []model.Message) string {
	// TODO: 实现更复杂的回复生成逻辑
	// 1. 可以接入OpenAI等AI服务
	// 2. 可以实现关键词匹配
	// 3. 可以实现意图识别
	// 4. 可以实现多轮对话管理

	return fmt.Sprintf("我收到了您的消息：%s\n如果您对服务满意，可以输入'feedback'进行评价。", text)
}
