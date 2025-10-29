package models

import (
	"log"
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

		case client := <-h.Unregister:
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				delete(h.UserClients, client.Username)
				close(client.Send)
				log.Printf("%s disconnected", client.Username)
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
