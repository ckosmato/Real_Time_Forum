package models

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Upgrader to upgrade HTTP connections to WebSocket
var Upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow connections from any origin
	},
}

// Message represents a chat message
type Message struct {
	Type      string `json:"type"`
	From      string `json:"from"`
	To        string `json:"to"`
	Content   string `json:"content"`
	Timestamp string `json:"timestamp"`
}

// Client represents a WebSocket client
type Client struct {
	Hub      *Hub
	Conn     *websocket.Conn
	Send     chan []byte
	Username string
}

// Hub maintains the set of active clients and broadcasts messages to the clients
type Hub struct {
	// Registered clients
	Clients map[*Client]bool

	// Inbound messages from the clients
	Broadcast chan []byte

	// Register requests from the clients
	Register chan *Client

	// Unregister requests from clients
	Unregister chan *Client

	// Map of usernames to clients for direct messaging
	UserClients map[string]*Client

	// Mutex for thread-safe operations
	Mutex sync.RWMutex
}

// NewHub creates a new Hub
func NewHub() *Hub {
	return &Hub{
		Broadcast:   make(chan []byte),
		Register:    make(chan *Client),
		Unregister:  make(chan *Client),
		Clients:     make(map[*Client]bool),
		UserClients: make(map[string]*Client),
	}
}

// Run starts the hub
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.Mutex.Lock()
			h.Clients[client] = true
			h.UserClients[client.Username] = client
			h.Mutex.Unlock()
			log.Printf("Client %s connected", client.Username)

		case client := <-h.Unregister:
			h.Mutex.Lock()
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				delete(h.UserClients, client.Username)
				close(client.Send)
			}
			h.Mutex.Unlock()
			log.Printf("Client %s disconnected", client.Username)

		case message := <-h.Broadcast:
			h.Mutex.RLock()
			for client := range h.Clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.Clients, client)
					delete(h.UserClients, client.Username)
				}
			}
			h.Mutex.RUnlock()
		}
	}
}

// SendToUser sends a message to a specific user
func (h *Hub) SendToUser(username string, message []byte) {
	h.Mutex.RLock()
	client, exists := h.UserClients[username]
	h.Mutex.RUnlock()

	if exists {
		select {
		case client.Send <- message:
		default:
			close(client.Send)
			h.Mutex.Lock()
			delete(h.Clients, client)
			delete(h.UserClients, username)
			h.Mutex.Unlock()
		}
	}
}

// GetActiveUsers returns a list of active users
func (h *Hub) GetActiveUsers() []string {
	h.Mutex.RLock()
	defer h.Mutex.RUnlock()

	users := make([]string, 0, len(h.UserClients))
	for username := range h.UserClients {
		users = append(users, username)
	}
	return users
}

// ReadPump pumps messages from the websocket connection to the hub
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	for {
		_, messageBytes, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		var msg Message
		if err := json.Unmarshal(messageBytes, &msg); err != nil {
			log.Printf("Error unmarshaling message: %v", err)
			continue
		}

		// Set the sender
		msg.From = c.Username

		// Set timestamp if not provided
		if msg.Timestamp == "" {
			msg.Timestamp = time.Now().Format(time.RFC3339)
		}

		// Save message to database if it's a chat message
		if msg.Type == "chat_message" && msg.To != "" && msg.To != "all" {
			chatMsg := &ChatMessage{
				FromUser:  msg.From,
				ToUser:    msg.To,
				Body:      msg.Content,
				CreatedAt: time.Now(),
			}

			// TODO: Save to database - this would need access to message repository
			// For now, we'll just log it
			log.Printf("Chat message from %s to %s: %s", chatMsg.FromUser, chatMsg.ToUser, chatMsg.Body)
		}

		// Marshal the message back to JSON
		messageBytes, err = json.Marshal(msg)
		if err != nil {
			log.Printf("Error marshaling message: %v", err)
			continue
		}

		// If it's a direct message, send to specific user
		if msg.To != "" && msg.To != "all" {
			c.Hub.SendToUser(msg.To, messageBytes)
			// Also send back to sender for confirmation
			select {
			case c.Send <- messageBytes:
			default:
				close(c.Send)
				return
			}
		} else {
			// Broadcast to all users
			c.Hub.Broadcast <- messageBytes
		}
	}
}

// WritePump pumps messages from the hub to the websocket connection
func (c *Client) WritePump() {
	defer c.Conn.Close()

	for message := range c.Send {
		if err := c.Conn.WriteMessage(websocket.TextMessage, message); err != nil {
			log.Printf("error: %v", err)
			return
		}
	}
	
	// Channel was closed, send close message
	c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
}
