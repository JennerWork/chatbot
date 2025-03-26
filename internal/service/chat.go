package service

import (
	"fmt"
	"strings"

	"github.com/JennerWork/chatbot/internal/model"
	"gorm.io/gorm"
)

// TextMessage represents the content of a text message
type TextMessage struct {
	Text string `json:"text"`
}

// ChatService handles the business logic related to chat
type ChatService interface {
	// ProcessText processes a text message and returns a reply
	ProcessText(customerID uint, sessionID uint, text string) (string, error)
}

type chatService struct {
	db *gorm.DB
}

// NewChatService creates an instance of the chat service
func NewChatService(db *gorm.DB) ChatService {
	return &chatService{
		db: db,
	}
}

// ProcessText processes a text message
func (s *chatService) ProcessText(customerID uint, sessionID uint, text string) (string, error) {
	// 1. Retrieve the context of the current session
	var session model.Session
	if err := s.db.First(&session, sessionID).Error; err != nil {
		return "", fmt.Errorf("failed to retrieve session: %v", err)
	}

	// 2. Check for any pending feedback
	var feedback model.Feedback
	result := s.db.Where("customer_id = ? AND status != ?",
		customerID, model.FeedbackStatusCompleted).
		Last(&feedback)
	hasPendingFeedback := result.Error == nil

	if hasPendingFeedback {
		// Process replies during the feedback process
		return s.handleFeedbackResponse(text, &feedback)
	}

	// 3. Check if the message is a feedback trigger
	if s.isFeedbackTrigger(text) {
		return s.initiateFeedback(customerID, sessionID)
	}

	// 4. Retrieve recent session messages for context understanding
	var recentMessages []model.Message
	if err := s.db.Where("session_id = ?", sessionID).
		Order("created_at desc").
		Find(&recentMessages).Error; err != nil {
		return "", fmt.Errorf("failed to retrieve historical messages: %v", err)
	}

	// 5. Generate a basic reply
	return s.generateBasicReply(text, recentMessages), nil
}

// isFeedbackTrigger checks if the message is a feedback trigger
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

// initiateFeedback initializes the feedback process
func (s *chatService) initiateFeedback(customerID uint, sessionID uint) (string, error) {
	// Create a new feedback record
	feedback := model.Feedback{
		CustomerID: customerID,
		SessionID:  sessionID,
		Status:     model.FeedbackStatusInitiated,
	}

	if err := s.db.Create(&feedback).Error; err != nil {
		return "", fmt.Errorf("failed to create feedback record: %v", err)
	}

	return "Please rate this service on a scale of 1-5, with 5 being very satisfied and 1 being very dissatisfied.", nil
}

// handleFeedbackResponse processes replies during the feedback process
func (s *chatService) handleFeedbackResponse(text string, feedback *model.Feedback) (string, error) {
	switch feedback.Status {
	case model.FeedbackStatusInitiated:
		// Process rating
		rating, err := s.parseRating(text)
		if err != nil {
			return "Please enter a number between 1 and 5 to rate.", nil
		}

		feedback.Rating = rating
		feedback.Status = model.FeedbackStatusRatingProvided
		if err := s.db.Save(feedback).Error; err != nil {
			return "", fmt.Errorf("failed to save rating: %v", err)
		}

		return "Thank you for your rating! Do you have any suggestions or comments about our service?", nil

	case model.FeedbackStatusRatingProvided:
		// Process comments
		sentimentService := NewSentimentAnalysisService()
		sentiment := sentimentService.AnalyzeSentiment(text, int(feedback.Rating))
		feedback.Comment = text
		feedback.Sentiment = sentiment
		feedback.Status = model.FeedbackStatusCompleted
		if err := s.db.Save(feedback).Error; err != nil {
			return "", fmt.Errorf("failed to save comment: %v", err)
		}

		return "Thank you very much for your feedback! We will continue to strive to provide better service.", nil

	default:
		return s.generateBasicReply(text, nil), nil
	}
}

// parseRating parses the rating
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
		return 0, fmt.Errorf("invalid rating")
	}
}

// generateBasicReply generates a basic reply
func (s *chatService) generateBasicReply(text string, context []model.Message) string {
	// TODO: Implement more complex reply generation logic
	// 1. Integrate with AI services like OpenAI
	// 2. Implement keyword matching
	// 3. Implement intent recognition
	// 4. Implement multi-turn conversation management

	return fmt.Sprintf("I have received your message: %s\nIf you are satisfied with the service, you can enter 'feedback' to provide a review.", text)
}
