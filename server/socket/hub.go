package socket

import (
	"context"
	"log/slog"
)

// Hub maintains the set of active clients and broadcasts messages to them.
// The Hub is the central component that manages all connected clients and message broadcasting.
// This approach encapsulates the concurrency logic for handling multiple clients.
type Hub struct {
	Broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	clients    map[*Client]bool
	logger     slog.Logger
}

func NewHub(logger slog.Logger) *Hub {
	return &Hub{
		logger:     logger,
		Broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

func (h *Hub) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			h.logger.Info("** WS hub shutting down **")
			// Gracefully close all client connections
			for client := range h.clients {
				close(client.send) // This will cause the writePump to exit
			}
			return ctx.Err() // Return the context's error (e.g., context.Canceled)
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case message := <-h.Broadcast:
			h.logger.Debug("broadcasting message", "message byte size", len(message))
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}
