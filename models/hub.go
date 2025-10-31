package models

import (
	"encoding/json"
	"log"
	"time"
)

func NewHub() *Hub {
	return &Hub{
		Clients:     make(map[*Client]bool),
		Register:    make(chan *Client),
		Unregister:  make(chan *Client),
		Broadcast:   make(chan []byte),
		UserClients: make(map[string]*Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.Clients[client] = true
			h.UserClients[client.Username] = client
			log.Printf("%s connected", client.Username)

			// Broadcast user join event
			h.broadcastUserStatusChange(client.Username, "user_joined")

		case client := <-h.Unregister:
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				delete(h.UserClients, client.Username)
				close(client.Send)
				log.Printf("%s disconnected", client.Username)

				// Broadcast user leave event
				h.broadcastUserStatusChange(client.Username, "user_left")
			}

		case message := <-h.Broadcast:
			for client := range h.Clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.Clients, client)
				}
			}
		}
	}
}

func (h *Hub) SendToUser(username string, message []byte) {
	if client, ok := h.UserClients[username]; ok {
		select {
		case client.Send <- message:
		default:
			close(client.Send)
			delete(h.Clients, client)
			delete(h.UserClients, username)
		}
	}
}

// broadcastUserStatusChange sends user join/leave notifications to all connected clients
func (h *Hub) broadcastUserStatusChange(username, eventType string) {
	statusMessage := Message{
		Type:      eventType,
		From:      "system",
		To:        "all",
		Content:   username,
		Timestamp: time.Now(),
	}

	// Send to all connected clients, but customize online users list for each
	for client := range h.Clients {
		// Get online users list excluding the current client
		onlineUsers := h.GetOnlineUsersExcluding(client.Username)

		messageData := map[string]interface{}{
			"type":         eventType,
			"from":         "system",
			"to":           "all",
			"content":      username,
			"timestamp":    statusMessage.Timestamp,
			"online_users": onlineUsers,
		}

		messageBytes, err := json.Marshal(messageData)
		if err != nil {
			log.Printf("Error marshaling status message: %v", err)
			continue
		}

		select {
		case client.Send <- messageBytes:
		default:
			close(client.Send)
			delete(h.Clients, client)
			delete(h.UserClients, client.Username)
		}
	}
}

// GetOnlineUsers returns a list of currently online users
func (h *Hub) GetOnlineUsers() []string {
	users := make([]string, 0, len(h.UserClients))
	for username := range h.UserClients {
		users = append(users, username)
	}
	return users
}

// GetOnlineUsersExcluding returns a list of currently online users excluding the specified username
func (h *Hub) GetOnlineUsersExcluding(excludeUsername string) []string {
	users := make([]string, 0, len(h.UserClients))
	for username := range h.UserClients {
		if username != excludeUsername {
			users = append(users, username)
		}
	}
	return users
}
