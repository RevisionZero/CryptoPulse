package hub

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/websocket"
)

// getAllowedOrigins returns the list of allowed origins from environment variable
// or defaults to localhost origins for development
func getAllowedOrigins() []string {
	originsEnv := os.Getenv("ALLOWED_ORIGINS")
	if originsEnv != "" {
		return strings.Split(originsEnv, ",")
	}
	// Default allowed origins for development
	return []string{
		"http://localhost:5173", // Vite default dev server
		"http://localhost:3000", // Common dev server port
		"http://127.0.0.1:5173",
		"http://127.0.0.1:3000",
	}
}

// checkOrigin validates the Origin header against the allowlist
func checkOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" {
		// No Origin header means it's not a browser request (e.g., direct connection)
		// Allow for testing purposes, but log it
		log.Printf("Warning: WebSocket connection without Origin header from %s", r.RemoteAddr)
		return true
	}

	allowedOrigins := getAllowedOrigins()
	for _, allowed := range allowedOrigins {
		allowed = strings.TrimSpace(allowed)
		if origin == allowed {
			return true
		}
	}

	log.Printf("Rejected WebSocket connection from unauthorized origin: %s", origin)
	return false
}

var upgrader = websocket.Upgrader{
	CheckOrigin:     checkOrigin,
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func (hub *Hub) WSHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
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
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		log.Printf("Received message from client: %s", msg)

		hub.symbolReqs <- SymbolRequest{
			Client:  conn,
			Symbols: strings.Split(string(msg), ","),
		}
	}
}
