package ws

import (
	"encoding/json"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer (64 KB).
	maxMessageSize = 64 * 1024

	// Size of the client send channel buffer.
	sendBufferSize = 256
)

// MessageHandler is a callback invoked when the client sends a WebSocket message.
// The hub passes the parsed Event and the sender's user ID.
type MessageHandler func(userID uuid.UUID, event Event)

// Client represents a single WebSocket connection for a user.
type Client struct {
	hub     *Hub
	conn    *websocket.Conn
	send    chan []byte
	UserID  uuid.UUID
	handler MessageHandler
}

// NewClient creates a new Client bound to the given Hub, connection, and user.
func NewClient(hub *Hub, conn *websocket.Conn, userID uuid.UUID, handler MessageHandler) *Client {
	return &Client{
		hub:     hub,
		conn:    conn,
		send:    make(chan []byte, sendBufferSize),
		UserID:  userID,
		handler: handler,
	}
}

// ReadPump pumps messages from the WebSocket connection to the hub.
// It runs in its own goroutine per client.
func (c *Client) ReadPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(maxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				log.Printf("ws client: unexpected close for user %s: %v", c.UserID, err)
			}
			break
		}

		var event Event
		if err := json.Unmarshal(message, &event); err != nil {
			log.Printf("ws client: invalid message from user %s: %v", c.UserID, err)
			continue
		}

		if c.handler != nil {
			c.handler(c.UserID, event)
		}
	}
}

// WritePump pumps messages from the hub to the WebSocket connection.
// It runs in its own goroutine per client.
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			if _, err := w.Write(message); err != nil {
				return
			}

			// Drain queued messages into the current write for efficiency.
			n := len(c.send)
			for i := 0; i < n; i++ {
				if _, err := w.Write([]byte("\n")); err != nil {
					break
				}
				if _, err := w.Write(<-c.send); err != nil {
					break
				}
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
