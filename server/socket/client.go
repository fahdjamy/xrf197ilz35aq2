package socket

import (
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"log/slog"
	"net/http"
	"time"
)

const (
	// Time allowed writing a message to the peer.
	writeWait = 5 * time.Second

	// Time allowed reading the next pong message from the peer.
	pongWait = 30 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

// The upgrader upgrades an HTTP connection to a WebSocket connection.
var upgrader = websocket.Upgrader{
	// ReadBufferSize and WriteBufferSize specify the I/O buffer size.
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// CheckOrigin returns true if the request Origin header is acceptable.
	// For production, implement a proper check here.
	// For DEV, we'll allow all connections.
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Client is an intermediary between the websocket connection and the hub.
type Client struct {
	hub  *Hub
	id   string
	send chan []byte
	log  slog.Logger
	conn *websocket.Conn
}

func NewClient(hub *Hub, conn *websocket.Conn, log slog.Logger) *Client {
	return &Client{
		hub:  hub,
		conn: conn,
		log:  log,
		id:   uuid.New().String(),
		send: make(chan []byte, 256),
	}
}

// readPump pumps messages from the websocket connection to the hub.
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		err := c.conn.Close()
		if err != nil {
			c.log.Error("close conn failed", "err", err)
		}
	}()
	c.conn.SetReadLimit(maxMessageSize)
	err := c.conn.SetReadDeadline(time.Now().Add(pongWait))
	if err != nil {
		return
	}
	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(pongWait))
	})
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseNormalClosure) {
				c.log.Error("unexpected close error", "err", err)
			}
			break
		}
		c.log.Info("client sent a message", "message", string(message), "id", c.id)
	}
}

// writePump pumps messages from the hub to the websocket connection.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		_ = c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			err := c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err != nil {
				return
			}
			if !ok {
				// The hub closed the channel.
				err = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				if err != nil {
					c.log.Error("write error", "err", err)
					return
				}
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				c.log.Error("next writer error", "err", err)
				return
			}
			_, err = w.Write(message)
			if err != nil {
				c.log.Error("write error", "err", err)
				return
			}

			if err := w.Close(); err != nil {
				c.log.Error("write error", "err", err)
				return
			}
		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				c.log.Error("write error", "err", err)
				return
			}
		}
	}
}
