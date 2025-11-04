package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"real-time-forum/models"
	"real-time-forum/services"
	"real-time-forum/utils"
	"strconv"
	"time"
)

type WebSocketHandler struct {
	chatService *services.ChatService
}

func NewWebSocketHandler(chatService *services.ChatService) *WebSocketHandler {
	return &WebSocketHandler{chatService: chatService}
}

// WebSocket upgrades the HTTP connection
func (h *WebSocketHandler) WebSocket(w http.ResponseWriter, r *http.Request) {
	user := utils.GetUserFromContext(r.Context())

	conn, err := models.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Failed to upgrade WebSocket", http.StatusInternalServerError)
		return
	}

	client := &models.Client{
		Username: user.Nickname,
		Conn:     conn,
		Send:     make(chan []byte, 256),
	}

	h.chatService.Hub.Register <- client

	// Send initial online users list after client is registered
	// Small delay to ensure registration is processed
	go func() {
		time.Sleep(100 * time.Millisecond)
		h.sendInitialOnlineUsers(client)
	}()

	// Start read/write pumps
	go h.readPump(client)
	go h.writePump(client)
}

// sendInitialOnlineUsers sends the current list of online users to a newly connected client
func (h *WebSocketHandler) sendInitialOnlineUsers(client *models.Client) {
	// Get online users excluding the current client
	onlineUsers := h.chatService.Hub.GetOnlineUsersExcluding(client.Username)

	initialMessage := map[string]interface{}{
		"type":         "initial_online_users",
		"from":         "system",
		"to":           client.Username,
		"online_users": onlineUsers,
		"timestamp":    fmt.Sprintf("%v", time.Now()),
	}

	messageBytes, err := json.Marshal(initialMessage)
	if err != nil {
		log.Printf("Error marshaling initial online users message: %v", err)
		return
	}

	// Send directly to the client's connection
	if err := client.Conn.WriteMessage(1, messageBytes); err != nil {
		log.Printf("Error sending initial online users: %v", err)
	}
}

func (h *WebSocketHandler) readPump(c *models.Client) {
	defer func() {
		h.chatService.Hub.Unregister <- c
		c.Conn.Close()
	}()

	for {
		_, msgBytes, err := c.Conn.ReadMessage()
		if err != nil {
			log.Printf("Read error: %v", err)
			break
		}

		var msg models.Message
		if err := json.Unmarshal(msgBytes, &msg); err != nil {
			log.Printf("JSON unmarshal error: %v", err)
			continue
		}

		msg.From = c.Username
		h.chatService.ProcessMessage(&msg)
	}
}

func (h *WebSocketHandler) writePump(c *models.Client) {
	defer c.Conn.Close()
	for msg := range c.Send {
		if err := c.Conn.WriteMessage(1, msg); err != nil {
			log.Printf("Write error: %v", err)
			break
		}
	}
}

func (h *WebSocketHandler) ChatHistory(w http.ResponseWriter, r *http.Request) {
	fmt.Print("ChatHistory handler called\n")
	query := r.URL.Query()
	user2 := query.Get("user2")
	limitStr := query.Get("limit")
	offsetStr := query.Get("offset")

	user := utils.GetUserFromContext(r.Context())

	// Default values for pagination
	limit := 10
	offset := 0

	// Parse pagination parameters
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 50 {
			limit = parsedLimit
		}
	}

	if offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	var history []models.Message
	var err error

	// Use pagination if offset is provided, otherwise use the old method for initial load
	if offset > 0 {
		history, err = h.chatService.GetChatHistoryWithPagination(r.Context(), user.Nickname, user2, limit, offset)
	} else {
		// For initial load, get the most recent messages
		history, err = h.chatService.GetChatHistoryWithPagination(r.Context(), user.Nickname, user2, limit, 0)
	}

	if err != nil {
		http.Error(w, "Failed to get chat history", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"history": history,
		"limit":   limit,
		"offset":  offset,
		"hasMore": len(history) == limit, // If we got exactly 'limit' messages, there might be more
	})
}
