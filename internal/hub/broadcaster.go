package hub

import "log"

// Broadcaster listens on the broadcast channel and sends messages to all connected clients
func (cm *ClientManager) Broadcaster() {
	log.Println("Broadcaster started")
	for message := range cm.broadcast {
		cm.SendToAll(message)
	}
}
