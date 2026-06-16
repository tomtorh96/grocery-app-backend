package ws

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

// Message is the structure sent over WebSocket to all clients
type Message struct {
	Type    string `json:"type"`
	Payload any    `json:"payload"`
}

// Client represents a single connected WebSocket client
type Client struct {
	conn   *websocket.Conn
	send   chan []byte
	listID string
	userID string
}

// Hub manages all connected clients grouped by list ID
type Hub struct {
	mu      sync.RWMutex
	clients map[string]map[*Client]bool // listID -> set of clients
}

// NewHub creates a new Hub
func NewHub() *Hub {
	return &Hub{
		clients: make(map[string]map[*Client]bool),
	}
}

// Register adds a client to a list's group
func (h *Hub) Register(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.clients[client.listID] == nil {
		h.clients[client.listID] = make(map[*Client]bool)
	}
	h.clients[client.listID][client] = true
}

// Unregister removes a client from a list's group
func (h *Hub) Unregister(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if clients, ok := h.clients[client.listID]; ok {
		delete(clients, client)
		if len(clients) == 0 {
			delete(h.clients, client.listID)
		}
	}
}

// Broadcast sends a message to all clients on a list except the sender
func (h *Hub) Broadcast(listID string, msg Message, exclude *Client) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("failed to marshal message: %v", err)
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for client := range h.clients[listID] {
		if client == exclude {
			continue
		}
		select {
		case client.send <- data:
		default:
			// client is too slow — drop the message
			log.Printf("client %s send buffer full, dropping message", client.userID)
		}
	}
}
