package models

import (
	"time"
)

// ChatMessage represents a chat message in the database
type ChatMessage struct {
	ID        int       `json:"id"`
	FromUser  string    `json:"from_user"`
	ToUser    string    `json:"to_user"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
}
