package socket

import (
	"log/slog"
	"net/http"
)

// ServeWS handles websocket requests from the peer.
func ServeWS(hub *Hub, w http.ResponseWriter, r *http.Request, log slog.Logger) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error("failed to upgrade WS connection", "err", err)
		return
	}
	client := NewClient(hub, conn, log)
	client.hub.register <- client

	go client.writePump()
	go client.readPump()
}
