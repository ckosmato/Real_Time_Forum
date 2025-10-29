package models

import (
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type Message struct {
	Type      string    `json:"type"`
	From      string    `json:"from"`
	To        string    `json:"to"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

type Client struct {
	ID       string
	Username string
	Conn     *websocket.Conn
	Send     chan []byte
}

type Hub struct {
	Clients     map[*Client]bool
	Register    chan *Client
	Unregister  chan *Client
	Broadcast   chan []byte
	UserClients map[string]*Client
}

var Upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true }, // allow any origin
}
