package services

import (
	"context"
	"encoding/json"
	"log"
	"real-time-forum/models"
	"real-time-forum/repositories"
	"sort"
	"time"
)

type ChatService struct {
	messageRepo *repositories.MessageRepository
	Hub         *models.Hub
}

func NewChatService(repo *repositories.MessageRepository, Hub *models.Hub) *ChatService {
	service := &ChatService{messageRepo: repo, Hub: Hub}

	// Set the user sorting function in the Hub
	Hub.SetUserSorter(service.SortUsersByLastMessage)

	return service
}

// Handle incoming message
func (s *ChatService) ProcessMessage(msg *models.Message) {
	msg.Timestamp = time.Now()

	// Use custom context to avoid cancellation
	dbCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := s.messageRepo.SaveMessage(dbCtx, msg)
	if err != nil {
		log.Printf("Error saving message: %v", err)
	}

	// Send to specific user
	messageBytes, _ := json.Marshal(msg)
	if msg.To != "" && msg.To != "all" {
		s.Hub.SendToUser(msg.To, messageBytes)
		// Also send back to sender
		s.Hub.SendToUser(msg.From, messageBytes)

		// Update online users list for both participants to reflect new conversation order
		s.refreshOnlineUsersOrder(msg.From, msg.To)
	} else {
		// Broadcast
		s.Hub.Broadcast <- messageBytes
	}
}

// refreshOnlineUsersOrder sends updated online users list to participants after a new message
func (s *ChatService) refreshOnlineUsersOrder(user1, user2 string) {
	// Send updated online users list to both participants
	for _, username := range []string{user1, user2} {
		onlineUsers := s.Hub.GetOnlineUsersExcluding(username)

		updateMessage := map[string]interface{}{
			"type":         "online_users_update",
			"from":         "system",
			"to":           username,
			"online_users": onlineUsers,
			"timestamp":    time.Now(),
		}

		messageBytes, err := json.Marshal(updateMessage)
		if err != nil {
			log.Printf("Error marshaling online users update: %v", err)
			continue
		}

		s.Hub.SendToUser(username, messageBytes)
	}
}

// Fetch chat history with pagination
func (s *ChatService) GetChatHistoryWithPagination(ctx context.Context, user1, user2 string, limit, offset int) ([]models.Message, error) {
	return s.messageRepo.GetMessagesWithPagination(ctx, user1, user2, limit, offset)
}

// SortUsersByLastMessage sorts users by putting those with recent conversations first, then alphabetically
func (s *ChatService) SortUsersByLastMessage(ctx context.Context, currentUser string, users []string) ([]string, error) {
	if len(users) == 0 {
		return users, nil
	}

	// Get last message timestamps for the current user
	timestamps, err := s.messageRepo.GetLastMessageTimestamps(ctx, currentUser)
	if err != nil {
		log.Printf("Error getting message timestamps for sorting: %v", err)
		// Fallback to alphabetical sorting
		sort.Strings(users)
		return users, nil
	}

	// Separate users with chat history from those without
	var usersWithChats []string
	var usersWithoutChats []string

	for _, user := range users {
		if _, hasChat := timestamps[user]; hasChat {
			usersWithChats = append(usersWithChats, user)
		} else {
			usersWithoutChats = append(usersWithoutChats, user)
		}
	}

	// Sort users with chats by timestamp (most recent first)
	sort.Slice(usersWithChats, func(i, j int) bool {
		timeI := timestamps[usersWithChats[i]]
		timeJ := timestamps[usersWithChats[j]]
		// More recent timestamp should come first (reverse chronological)
		return timeI > timeJ
	})

	// Sort users without chats alphabetically
	sort.Strings(usersWithoutChats)

	// Combine: users with chats first, then users without chats
	result := make([]string, 0, len(users))
	result = append(result, usersWithChats...)
	result = append(result, usersWithoutChats...)

	return result, nil
}
