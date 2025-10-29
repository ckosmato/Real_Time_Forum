package services

import (
	"context"
	"encoding/json"
	"log"
	"real-time-forum/models"
	"real-time-forum/repositories"
	"time"
)

type ChatService struct {
	messageRepo *repositories.MessageRepository
	Hub         *models.Hub
}

func NewChatService(repo *repositories.MessageRepository, Hub *models.Hub) *ChatService {
	return &ChatService{messageRepo: repo, Hub: Hub}
}

// Handle incoming message
func (s *ChatService) ProcessMessage(ctx context.Context, msg *models.Message) {
	msg.Timestamp = time.Now()

	// Save message to DB
	err := s.messageRepo.SaveMessage(ctx, &models.ChatMessage{
		FromUser:  msg.From,
		ToUser:    msg.To,
		Body:      msg.Content,
		CreatedAt: msg.Timestamp,
	})
	if err != nil {
		log.Printf("Error saving message: %v", err)
	}

	// Send to specific user
	messageBytes, _ := json.Marshal(msg)
	if msg.To != "" && msg.To != "all" {
		s.Hub.SendToUser(msg.To, messageBytes)
		// Also send back to sender
		s.Hub.SendToUser(msg.From, messageBytes)
	} else {
		// Broadcast
		s.Hub.Broadcast <- messageBytes
	}
}
