package hub

import (
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins in development
	},
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
