package handlers

import (
	"encoding/json"
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

// HandleWebSocket handles WebSocket connections
func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	user := utils.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

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

// GetOnlineUsers returns the list of currently online users
func GetOnlineUsers(w http.ResponseWriter, r *http.Request) {
	if Hub == nil {
		http.Error(w, "Hub not initialized", http.StatusInternalServerError)
		return
	}

	users := Hub.GetActiveUsers()

	response := struct {
		Users []string `json:"users"`
	}{
		Users: users,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
	}
}
