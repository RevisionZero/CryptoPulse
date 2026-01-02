package hub

import (
	"log"
	"net/http"
	"regexp"
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

const (
	maxSymbolLength = 20   // Maximum length for a single symbol
	maxSymbolsCount = 100  // Maximum number of symbols per request
)

// symbolPattern validates that symbols contain only uppercase letters and numbers
// Binance symbols are typically like BTCUSDT, ETHUSDT, etc.
var symbolPattern = regexp.MustCompile(`^[A-Z0-9]+$`)

// sanitizeForLog limits and sanitizes strings for safe logging
func sanitizeForLog(s string, maxLen int) string {
	if len(s) > maxLen {
		s = s[:maxLen]
	}
	// Replace any non-printable characters with '?'
	return strings.Map(func(r rune) rune {
		if r < 32 || r > 126 {
			return '?'
		}
		return r
	}, s)
}


// validateSymbols validates and sanitizes the symbols received from clients
// Returns only valid symbols, filtering out malformed or malicious inputs
func validateSymbols(symbols []string) []string {
	if len(symbols) > maxSymbolsCount {
		log.Printf("Warning: Too many symbols requested (%d), limiting to %d", len(symbols), maxSymbolsCount)
		symbols = symbols[:maxSymbolsCount]
	}

	validSymbols := make([]string, 0, len(symbols))
	for _, symbol := range symbols {
		// Trim whitespace
		symbol = strings.TrimSpace(symbol)

		// Skip empty symbols
		if symbol == "" {
			continue
		}

		// Check length
		if len(symbol) > maxSymbolLength {
			log.Printf("Warning: Symbol too long (length %d): %s", len(symbol), sanitizeForLog(symbol, 30))
			continue
		}

		// Validate pattern (uppercase alphanumeric only)
		if !symbolPattern.MatchString(symbol) {
			log.Printf("Warning: Invalid symbol format: %s", sanitizeForLog(symbol, 30))
			continue
		}

		validSymbols = append(validSymbols, symbol)
	}

	return validSymbols
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

		log.Printf("Received message from client: %s", sanitizeForLog(string(msg), 200))

		// Validate and sanitize symbols before processing
		rawSymbols := strings.Split(string(msg), ",")
		validSymbols := validateSymbols(rawSymbols)

		// Only process if we have valid symbols
		if len(validSymbols) == 0 {
			log.Printf("Warning: No valid symbols in request from client")
			continue
		}

		hub.symbolReqs <- SymbolRequest{
			Client:  conn,
			Symbols: validSymbols,
		}
	}
}
