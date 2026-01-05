package hub

import (
	"bytes"
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
	Send      chan []byte // Channel to push data to this specific client
	Conn      bool
	PCCMatrix map[string]map[string]float64
}

type SymbolRequest struct {
	Client  *websocket.Conn
	Symbols []string
}

type Hub struct {
	clients    map[*websocket.Conn]*Client
	symbols    map[string]*models.SymbolAttributes
	symbolLock sync.Mutex
	Connect    chan *websocket.Conn
	Disconnect chan *websocket.Conn
	symbolReqs chan SymbolRequest
	broadcast  chan map[string][]float64
}

var bufferPool = sync.Pool{
	New: func() interface{} {
		// This is called if the responses pool is empty
		return new(bytes.Buffer)
	},
}

func NewHub(broadcast chan map[string][]float64) *Hub {
	return &Hub{
		clients:    make(map[*websocket.Conn]*Client),
		symbols:    make(map[string]*models.SymbolAttributes),
		broadcast:  broadcast,
		Connect:    make(chan *websocket.Conn, 64),
		Disconnect: make(chan *websocket.Conn, 128),
		symbolReqs: make(chan SymbolRequest, 64),
	}
}

func (hub *Hub) Run() {

	const channelCapacity = 100
	rawData := make(chan []byte, channelCapacity)

	for symbol, _ := range hub.symbols {

		go connection.Connector([]string{symbol}, rawData)

	}

	go engine.Synchronizer(hub.symbols, rawData, hub.broadcast, &hub.symbolLock)
	for {
		select {
		case message := <-hub.broadcast:
			hub.SendToAll(message)
		default:
			// Put justification here for why putting the case again
			select {
			case message := <-hub.broadcast:
				hub.SendToAll(message)
			case conn := <-hub.Connect:
				hub.AddClient(conn)
			case conn := <-hub.Disconnect:
				hub.RemoveClient(conn)
			case symbolRequest := <-hub.symbolReqs:
				hub.HandleSymbolRequest(symbolRequest, rawData)
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

func (hub *Hub) HandleSymbolRequest(symbolRequest SymbolRequest, dataStream chan []byte) {
	log.Print("Symbols requested: ", symbolRequest.Symbols)
	client := hub.clients[symbolRequest.Client]
	client.Symbols = []string{}
	for _, symbol := range symbolRequest.Symbols {
		log.Print("Symbol requested: ", symbol)
		hub.symbolLock.Lock()
		if _, exists := hub.symbols[symbol]; !exists {
			hub.symbols[symbol] = &models.SymbolAttributes{
				LatestPrice:   0.0,
				SlidingWindow: utils.NewRingBuffer(600),
			}
			hub.symbolLock.Unlock()
			go connection.Connector([]string{symbol}, dataStream)
		} else {
			hub.symbolLock.Unlock()
		}

		client.Symbols = append(client.Symbols, symbol)
	}
	sort.Strings(client.Symbols)
	ModifyClientMatrix(client)
}

func (hub *Hub) AddClient(conn *websocket.Conn) {
	// hub.mu.Lock()
	// defer hub.mu.Unlock()
	hub.clients[conn] = &Client{
		ID:        conn,
		Symbols:   []string{},
		Send:      make(chan []byte, 30),
		Conn:      true,
		PCCMatrix: make(map[string]map[string]float64, 0),
	}
	go hub.clients[conn].writePump(hub)
	log.Printf("Client connected. Total clients: %d", len(hub.clients))
}

func (hub *Hub) RemoveClient(conn *websocket.Conn) {
	if _, ok := hub.clients[conn]; ok {
		close(hub.clients[conn].Send)
		delete(hub.clients, conn)
		conn.Close()
		log.Printf("Client disconnected. Total clients: %d", len(hub.clients))
	}
}

func (hub *Hub) SendToAll(sampledData map[string][]float64) {
	// hub.mu.Lock()
	// defer hub.mu.Unlock()

	// All responses to all clients, a map with the key being the symbols
	// requested. This is to minimize the amount of PCC calculations done,
	// as many clients might share identical matrices

	// responses := make(map[string]map[string]map[string]float64)
	responses := make(map[string][]byte)

	for _, client := range hub.clients {
		key := strings.Join(client.Symbols, ",")
		jsonData, exists := responses[key]
		if !exists {
			// 1. Get a buffer from the pool
			buf := bufferPool.Get().(*bytes.Buffer)
			buf.Reset() // CRITICAL: Clear any data from previous use
			engine.CalculatePCCMatrix(sampledData, client.Symbols, client.PCCMatrix)
			jsonErr := json.NewEncoder(buf).Encode(client.PCCMatrix)

			if jsonErr != nil {
				log.Printf("JSON Encode Error: %v", jsonErr)
				bufferPool.Put(buf) // Return even on error
				continue
			}
			// responses[key], _ = json.Marshal(client.PCCMatrix)
			jsonData = append([]byte(nil), buf.Bytes()...)
			responses[key] = jsonData
			bufferPool.Put(buf) // Return buffer to pool
		}
		select {
		case client.Send <- jsonData:
			// Message sent successfully
		default:
			// Skip if buffer is full or channel is closed to keep Hub fast
			log.Printf("Skipping slow client:")
		}
	}

}

func (c *Client) writePump(hub *Hub) {
	defer func() {
		select {
		case hub.Disconnect <- c.ID:
			// Signal sent successfully
		default:
			// Signal already sent by the other pump, or Hub is busy
		}
	}()

	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				// The Hub closed the channel, send a close message to client
				c.ID.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// Perform the actual network write
			err := c.ID.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				log.Printf("Write error for client %v: %v", c.ID.RemoteAddr(), err)
				return
			}
		}
	}
}
