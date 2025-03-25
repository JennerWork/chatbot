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

type SessionStatus string

const (
	SessionStatusInitiated SessionStatus = "initiated"
	SessionStatusActive    SessionStatus = "active"
	SessionStatusCancelled SessionStatus = "cancelled"
	SessionStatusTimedOut  SessionStatus = "timed_out"
)
