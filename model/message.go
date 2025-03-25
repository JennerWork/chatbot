package model

import "gorm.io/gorm"

type Message struct {
	gorm.Model
	Content    string   `json:"content"`
	CustomerID uint     `json:"customer_id"`
	Customer   Customer `json:"customer"`
	Sender     Sender   `json:"sender"`
	SessionID  uint     `json:"session_id"`
	Session    Session  `json:"session"`
}

type Sender string

const (
	SenderCustomer Sender = "customer"
	SenderBot      Sender = "bot"
)
