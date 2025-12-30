package hub

import (
	"encoding/json"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

type ClientManager struct {
	clients   map[*websocket.Conn]bool
	mu        sync.Mutex
	broadcast chan map[string]map[string]float64
}

func NewClientManager() *ClientManager {
	return &ClientManager{
		clients:   make(map[*websocket.Conn]bool),
		broadcast: make(chan map[string]map[string]float64, 256),
	}
}

func (cm *ClientManager) AddClient(conn *websocket.Conn) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.clients[conn] = true
	log.Printf("Client connected. Total clients: %d", len(cm.clients))
}

func (cm *ClientManager) RemoveClient(conn *websocket.Conn) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	if _, ok := cm.clients[conn]; ok {
		delete(cm.clients, conn)
		conn.Close()
		log.Printf("Client disconnected. Total clients: %d", len(cm.clients))
	}
}

func (cm *ClientManager) Broadcast(message map[string]map[string]float64) {
	cm.broadcast <- message
}

func (cm *ClientManager) GetBroadcastChannel() chan map[string]map[string]float64 {
	return cm.broadcast
}

func (cm *ClientManager) SendToAll(message map[string]map[string]float64) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Convert map to JSON bytes
	jsonData, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return
	}

	for client := range cm.clients {
		err := client.WriteMessage(websocket.TextMessage, jsonData)
		if err != nil {
			log.Printf("Error sending to client: %v", err)
			client.Close()
			delete(cm.clients, client)
		}
	}
}
