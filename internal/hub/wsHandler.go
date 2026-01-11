package hub

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
)

// var upgrader = websocket.Upgrader{
// 	CheckOrigin: func(r *http.Request) bool {
// 		return true // Allow all origins in development
// 	},
// 	ReadBufferSize:  1024,
// 	WriteBufferSize: 1024,
// }

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Retrieve the 'Origin' header from the incoming request
		origin := r.Header.Get("Origin")

		// Only allow your production domain (and localhost for testing)
		// Ensure you include the protocol (https://)
		allowedOrigins := []string{
			"https://cryptopulseapp.dev",
			"http://localhost:5173", // Optional: Keep Vite's dev port for local testing
		}

		for _, allowed := range allowedOrigins {
			if origin == allowed {
				return true
			}
		}
		return false // Block all other origins
	},
}

func (hub *Hub) WSHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Info("WebSocket upgrade error: %v", err)
		return
	}

	// Register client
	hub.Connect <- conn

	// Cleanup on disconnect
	// defer hub.RemoveClient(conn)
	defer func() {
		select {
		case hub.Disconnect <- conn:
			// Signal sent successfully
		default:
			// Signal already sent by the other pump, or Hub is busy
		}
	}()

	// Keep connection alive - read messages from client
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				slog.Info("WebSocket error: %v", err)
			}
			break
		}

		slog.Info("Received message from client: %s", msg)

		hub.symbolReqs <- SymbolRequest{
			Client:  conn,
			Symbols: strings.Split(string(msg), ","),
		}
	}
}
