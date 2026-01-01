package hub

import "log"

// Broadcaster listens on the broadcast channel and sends messages to all connected clients
func (hub *Hub) Broadcaster() {
	log.Println("Broadcaster started")
	for sampledData := range hub.broadcast {
		hub.SendToAll(sampledData)
	}
}
