package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"real-time-forum/models"
	"real-time-forum/services"
	"real-time-forum/utils"
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

	// Send initial online users list to the newly connected client
	h.sendInitialOnlineUsers(client)

	h.chatService.Hub.Register <- client

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
	user := utils.GetUserFromContext(r.Context())
	//fmt.Println("Fetching chat history between:", user.Nickname, "and", user2)
	history, err := h.chatService.GetChatHistory(r.Context(), user.Nickname, user2)
	if err != nil {
		http.Error(w, "Failed to get chat history", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"history": history,
	})
}
