package model

import "gorm.io/gorm"

type Feedback struct {
	gorm.Model
	Status     FeedbackStatus `json:"status"`
	CustomerID uint           `json:"customer_id"`
	Customer   Customer       `json:"customer"`
	Rating     FeedbackRating `json:"rating"`
	MessageID  uint           `json:"message_id"`
	Message    Message        `json:"message"`
	SessionID  uint           `json:"session_id"`
	Session    Session        `json:"session"`
	Comment    string         `json:"comment"`
}

type FeedbackRating int

const (
	FeedbackRating1 FeedbackRating = 1
	FeedbackRating2 FeedbackRating = 2
	FeedbackRating3 FeedbackRating = 3
	FeedbackRating4 FeedbackRating = 4
	FeedbackRating5 FeedbackRating = 5
)

type FeedbackStatus string

const (
	FeedbackStatusInitiated       FeedbackStatus = "initiated"
	FeedbackStatusRatingProvided  FeedbackStatus = "rating_provided"
	FeedbackStatusCommentProvided FeedbackStatus = "comment_provided"
	FeedbackStatusCompleted       FeedbackStatus = "completed"
	FeedbackStatusTimeout         FeedbackStatus = "timeout"
	FeedbackStatusCancelled       FeedbackStatus = "cancelled"
)
