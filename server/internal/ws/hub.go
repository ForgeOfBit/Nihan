package ws

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/google/uuid"
)

// Event represents a WebSocket message exchanged between client and server.
type Event struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// Hub maintains the set of active clients and broadcasts messages to them.
type Hub struct {
	// clients maps user ID to their active client connection.
	clients map[uuid.UUID]*Client

	// mu protects the clients map for concurrent access.
	mu sync.RWMutex

	// register channel for new client connections.
	register chan *Client

	// unregister channel for disconnecting clients.
	unregister chan *Client

	// stop channel signals the Hub to shut down.
	stop chan struct{}
}

// NewHub creates and returns a new Hub instance.
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[uuid.UUID]*Client),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		stop:       make(chan struct{}),
	}
}

// Run starts the Hub's main event loop. It must be called in a goroutine.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			// Close any existing connection for the same user (single session).
			if existing, ok := h.clients[client.UserID]; ok {
				close(existing.send)
				delete(h.clients, client.UserID)
			}
			h.clients[client.UserID] = client
			h.mu.Unlock()
			log.Printf("ws hub: user %s connected (total: %d)", client.UserID, h.ClientCount())

		case client := <-h.unregister:
			h.mu.Lock()
			if existing, ok := h.clients[client.UserID]; ok && existing == client {
				close(existing.send)
				delete(h.clients, client.UserID)
			}
			h.mu.Unlock()
			log.Printf("ws hub: user %s disconnected (total: %d)", client.UserID, h.ClientCount())

		case <-h.stop:
			h.mu.Lock()
			for id, client := range h.clients {
				close(client.send)
				delete(h.clients, id)
			}
			h.mu.Unlock()
			return
		}
	}
}

// Stop signals the Hub to shut down and clean up all connections.
func (h *Hub) Stop() {
	close(h.stop)
}

// SendToUser sends an event to a specific user if they are online.
// Returns true if the message was sent, false if the user is not connected.
func (h *Hub) SendToUser(userID uuid.UUID, event Event) bool {
	h.mu.RLock()
	client, ok := h.clients[userID]
	h.mu.RUnlock()

	if !ok {
		return false
	}

	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("ws hub: failed to marshal event for user %s: %v", userID, err)
		return false
	}

	select {
	case client.send <- data:
		return true
	default:
		// Buffer full – drop the message to avoid blocking.
		log.Printf("ws hub: send buffer full for user %s, dropping message", userID)
		return false
	}
}

// IsOnline checks if a user is currently connected.
func (h *Hub) IsOnline(userID uuid.UUID) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, ok := h.clients[userID]
	return ok
}

// Register adds a new client to the hub via the register channel.
func (h *Hub) Register(client *Client) {
	h.register <- client
}

// ClientCount returns the number of currently connected clients.
func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}
