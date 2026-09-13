package hub

import (
	"bytes"
	"encoding/json"
	"log"
	"log/slog"
	"main/internal/connection"
	"main/internal/engine"
	"main/pkg/models"
	"main/pkg/utils"
	"sort"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

// Client tracks one websocket peer and the symbols/matrix currently associated with it.
type Client struct {
	ID        *websocket.Conn
	Symbols   []string    // The coins they currently want
	Send      chan []byte // Channel to push data to this specific client
	Conn      bool
	PCCMatrix map[string]map[string]float64
}

// SymbolRequest represents one inbound symbol subscription update from a client.
type SymbolRequest struct {
	Client  *websocket.Conn
	Symbols []string
}

// Hub coordinates websocket clients, symbol lifecycle, and computed broadcast payloads.
type Hub struct {
	clients    map[*websocket.Conn]*Client
	symbols    map[string]*models.SymbolAttributes
	symbolLock sync.Mutex
	Connect    chan *websocket.Conn
	Disconnect chan *websocket.Conn
	symbolReqs chan SymbolRequest
	broadcast  chan map[string][]float64
}

// bufferPool reuses temporary encoding buffers to reduce allocations during fan-out.
var bufferPool = sync.Pool{
	New: func() interface{} {
		// This is called if the responses pool is empty
		return new(bytes.Buffer)
	},
}

// NewHub builds a hub with pre-sized channels for expected connect/disconnect bursts.
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

// Run starts the symbol synchronizer and processes all hub events in a single loop.
func (hub *Hub) Run() {

	// Keep a modest queue so short read spikes from exchange streams do not block producers.
	const channelCapacity = 100
	rawData := make(chan []byte, channelCapacity)

	go engine.Synchronizer(hub.symbols, rawData, hub.broadcast, &hub.symbolLock)
	for {
		select {
		case message := <-hub.broadcast:
			hub.SendToAll(message)
		default:
			// Put justification here for why putting the case again
			// The nested select gives connection management a chance to run when no broadcast is ready.
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

// ModifyClientMatrix initializes/refreshes a client's PCC matrix for the active symbol set.
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

// HandleSymbolRequest updates hub symbol state and starts upstream streams as symbols appear.
func (hub *Hub) HandleSymbolRequest(symbolRequest SymbolRequest, dataStream chan []byte) {
	log.Print("Symbols requested: ", symbolRequest.Symbols)
	client := hub.clients[symbolRequest.Client]
	client.Symbols = []string{}
	for _, symbol := range symbolRequest.Symbols {
		log.Print("Symbol requested: ", symbol)
		hub.symbolLock.Lock()
		if _, exists := hub.symbols[symbol]; !exists {
			hub.symbols[symbol] = &models.SymbolAttributes{
				LatestPrice: 0.0,
				// Keep 600 samples (~60s at 100ms sampling cadence) for PCC calculations.
				SlidingWindow: utils.NewRingBuffer(600),
				ClientCounter: 1,
				Close:         make(chan bool, 1),
			}
			// Connector expects a symbol list; this stream is intentionally scoped to one symbol.
			go connection.Connector([]string{symbol}, dataStream, hub.symbols[symbol].Close)
			hub.symbolLock.Unlock()
		} else {
			hub.symbols[symbol].ClientCounter++
			hub.symbolLock.Unlock()
		}

		client.Symbols = append(client.Symbols, symbol)
	}
	sort.Strings(client.Symbols)
	ModifyClientMatrix(client)
}

// AddClient registers a websocket and starts its dedicated write pump.
func (hub *Hub) AddClient(conn *websocket.Conn) {
	// hub.mu.Lock()
	// defer hub.mu.Unlock()
	hub.clients[conn] = &Client{
		ID:      conn,
		Symbols: []string{},
		// Buffer outbound messages so occasional slow writes do not stall hub event handling.
		Send:      make(chan []byte, 30),
		Conn:      true,
		PCCMatrix: make(map[string]map[string]float64, 0),
	}
	go hub.clients[conn].writePump(hub)
	slog.Info("Client connected. Total clients: %d", len(hub.clients))
}

// RemoveClient removes a websocket client and releases symbol resources it no longer needs.
func (hub *Hub) RemoveClient(conn *websocket.Conn) {
	if _, ok := hub.clients[conn]; ok {
		hub.RemoveSymbols(conn)
		close(hub.clients[conn].Send)
		delete(hub.clients, conn)
		conn.Close()
		slog.Info("Client disconnected. Total clients: %d", len(hub.clients))
	}
}

// RemoveSymbols decrements symbol subscriber counts and tears down orphaned exchange streams.
func (hub *Hub) RemoveSymbols(conn *websocket.Conn) {
	for _, symbol := range hub.clients[conn].Symbols {
		hub.symbolLock.Lock()
		if attr, exists := hub.symbols[symbol]; exists {
			attr.ClientCounter--
			if attr.ClientCounter <= 0 {
				attr.Close <- true
				delete(hub.symbols, symbol)
				slog.Info("Deleted symbol: ", symbol)
			}
		}
		hub.symbolLock.Unlock()
	}

}

// SendToAll computes/reuses PCC payloads and fan-outs messages to each connected client.
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
				slog.Info("JSON Encode Error: %v", jsonErr)
				bufferPool.Put(buf) // Return even on error
				continue
			}
			// responses[key], _ = json.Marshal(client.PCCMatrix)
			// Copy bytes so pooled buffer reuse cannot mutate data already queued for clients.
			jsonData = append([]byte(nil), buf.Bytes()...)
			responses[key] = jsonData
			bufferPool.Put(buf) // Return buffer to pool
		}
		select {
		case client.Send <- jsonData:
			// Message sent successfully
		default:
			// Skip if buffer is full or channel is closed to keep Hub fast
			slog.Info("Skipping slow client:")
		}
	}

}

// writePump serializes outbound websocket writes for a single client connection.
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
				slog.Info("Write error for client %v: %v", c.ID.RemoteAddr(), err)
				return
			}
		}
	}
}
