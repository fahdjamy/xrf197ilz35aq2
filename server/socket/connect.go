package socket

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func StartWSServer(log slog.Logger, hub *Hub) {
	http.HandleFunc("/xrf-ws", func(w http.ResponseWriter, r *http.Request) {
		// TODO: Authenticate the user here before upgrading the connection.
		ServeWS(hub, w, r, log)
	})
	// TODO: IN production, use ListenAndServeTLS
	server := &http.Server{
		Addr: ":8082",
	}

	go func() {
		log.Info("starting ws http server on port 8082")
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("starting http server error", err)
			return
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("shutting down WS server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Error("WS server shutdown error", "err", err.Error())
	}
	log.Info("WS server gracefully stopped")
}

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
