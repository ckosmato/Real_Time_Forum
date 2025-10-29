package handlers

import (
	"log"
	"net/http"
	"real-time-forum/models"
	"real-time-forum/utils"
)

// Global hub instance
var Hub *models.Hub

// InitHub initializes the WebSocket hub
func InitHub() {
	Hub = models.NewHub()
	go Hub.Run()
}
type WebSocketHandler struct{}

func NewWebSocketHandler() *WebSocketHandler {
	return &WebSocketHandler{}
}
// WebSocket handles WebSocket connections
func (h *WebSocketHandler) WebSocket(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	user := utils.GetUserFromContext(r.Context())

	// Upgrade the HTTP connection to WebSocket
	conn, err := models.Upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}

	// Create a new client
	client := &models.Client{
		Hub:      Hub,
		Conn:     conn,
		Send:     make(chan []byte, 256),
		Username: user.Nickname,
	}

	// Register the client
	client.Hub.Register <- client

	// Start goroutines for reading and writing
	go client.WritePump()
	go client.ReadPump()
}

