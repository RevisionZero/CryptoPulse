package hub

import (
	"encoding/json"
	"log"
	"main/internal/connection"
	"main/internal/engine"
	"main/pkg/models"
	"main/pkg/utils"
	"sort"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

type Client struct {
	ID        *websocket.Conn
	Symbols   []string    // The coins they currently want
	Send      chan string // Channel to push data to this specific client
	Receive   chan string // Channel to receive data from this specific client
	Conn      bool
	PCCMatrix map[string]map[string]float64
}

type SymbolRequest struct {
	Client  *Client
	Symbols []string
}

type Hub struct {
	clients    map[*websocket.Conn]*Client
	symbols    map[string]*models.SymbolAttributes
	connect    chan *websocket.Conn
	disconnect chan *websocket.Conn
	symbolReqs chan SymbolRequest
	mu         sync.Mutex
	broadcast  chan map[string][]float64
}

func NewHub(broadcast chan map[string][]float64) *Hub {
	return &Hub{
		clients:    make(map[*websocket.Conn]*Client),
		symbols:    make(map[string]*models.SymbolAttributes),
		broadcast:  broadcast,
		connect:    make(chan *websocket.Conn),
		disconnect: make(chan *websocket.Conn),
		symbolReqs: make(chan SymbolRequest),
	}
}

func (hub *Hub) Run() {

	const channelCapacity = 100
	rawData := make(chan []byte, channelCapacity)

	for symbol, _ := range hub.symbols {

		go connection.Connector([]string{symbol}, rawData)

	}

	go engine.Synchronizer(hub.symbols, rawData, hub.broadcast)
	for {
		select {
		case message := <-hub.broadcast:
			hub.SendToAll(message)
		default:
			// Put justification here for why putting the case again
			select {
			case message := <-hub.broadcast:
				hub.SendToAll(message)
			case conn := <-hub.connect:
				hub.AddClient(conn)
			case conn := <-hub.disconnect:
				hub.RemoveClient(conn)
			case symbolRequest := <-hub.symbolReqs:
				log.Print("Symbols requested: ", symbolRequest.Symbols)
				symbolRequest.Client.Symbols = []string{}
				for _, symbol := range symbolRequest.Symbols {
					log.Print("Symbol requested: ", symbol)
					if _, exists := hub.symbols[symbol]; !exists {
						hub.symbols[symbol] = &models.SymbolAttributes{
							LatestPrice:   0.0,
							SlidingWindow: utils.NewRingBuffer(600),
						}
						go connection.Connector([]string{symbol}, rawData)
					}
					symbolRequest.Client.Symbols = append(symbolRequest.Client.Symbols, symbol)
					// sort.Strings(symbolRequest.Client.Symbols)
					// ModifyClientMatrix(symbolRequest.Client)
				}
				sort.Strings(symbolRequest.Client.Symbols)
				ModifyClientMatrix(symbolRequest.Client)
			}

		}

	}
}

func ModifyClientMatrix(client *Client) {
	pccMatrix := make(map[string]map[string]float64, len(client.Symbols))
	for _, symbolX := range client.Symbols {
		pccMatrix[symbolX] = make(map[string]float64, len(client.Symbols))
		for _, symbolY := range client.Symbols {
			if symbolX == symbolY {
				pccMatrix[symbolX][symbolY] = 1.0
			}
		}
	}
	client.PCCMatrix = pccMatrix
}

func (hub *Hub) AddClient(conn *websocket.Conn) {
	hub.mu.Lock()
	defer hub.mu.Unlock()
	hub.clients[conn] = &Client{
		ID:        conn,
		Symbols:   []string{},
		Send:      make(chan string, 10),
		Receive:   make(chan string, 1),
		Conn:      true,
		PCCMatrix: make(map[string]map[string]float64, 0),
	}
	log.Printf("Client connected. Total clients: %d", len(hub.clients))
}

func (hub *Hub) RemoveClient(conn *websocket.Conn) {
	hub.mu.Lock()
	defer hub.mu.Unlock()
	if _, ok := hub.clients[conn]; ok {
		delete(hub.clients, conn)
		conn.Close()
		log.Printf("Client disconnected. Total clients: %d", len(hub.clients))
	}
}

func (hub *Hub) Broadcast(message map[string][]float64) {
	hub.broadcast <- message
}

func (hub *Hub) GetBroadcastChannel() chan map[string][]float64 {
	return hub.broadcast
}

func (hub *Hub) SendToAll(sampledData map[string][]float64) {
	hub.mu.Lock()
	defer hub.mu.Unlock()

	// All responses to all clients, a map with the key being the symbols
	// requested. This is to minimize the amount of PCC calculations done,
	// as many clients might share identical matrices
	responses := make(map[string]map[string]map[string]float64)

	for _, client := range hub.clients {
		key := strings.Join(client.Symbols, ",")
		if _, exists := responses[key]; !exists {
			engine.CalculatePCCMatrix(sampledData, client.Symbols, client.PCCMatrix)
			responses[key] = client.PCCMatrix
		}
		jsonData, marshalErr := json.Marshal(responses[key])
		if marshalErr != nil {
			log.Printf("Error marshaling message: %v", marshalErr)
			continue
		}
		sendErr := client.ID.WriteMessage(websocket.TextMessage, jsonData)
		if sendErr != nil {
			log.Printf("Error sending to client: %v", sendErr)
			client.ID.Close()
			delete(hub.clients, client.ID)
			hub.RemoveClient(client.ID)
		}
	}

}
