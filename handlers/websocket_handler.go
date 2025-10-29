package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"real-time-forum/models"
	"real-time-forum/services"
	"real-time-forum/utils"
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

	// Start read/write pumps
	go h.readPump(client, r.Context())
	go h.writePump(client)
}

func (h *WebSocketHandler) readPump(c *models.Client, ctx context.Context) {
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
		h.chatService.ProcessMessage(ctx, &msg)
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
