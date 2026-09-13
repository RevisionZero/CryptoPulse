package connection

import (
	"fmt"
	"os"
	"strings"
)

// constructBinanceEndpoint builds the combined-stream websocket URL for the given symbols.
func constructBinanceEndpoint(symbols []string) string {
	streamsURL := ""
	for _, symbol := range symbols {
		streamsURL += fmt.Sprintf("%s@bookTicker/", strings.ToLower(symbol))
	}
	// Trim the final "/" so the stream list is valid for Binance's query format.
	streamsURL = streamsURL[:len(streamsURL)-1]

	baseURL := os.Getenv("BINANCE_WS_URL")
	if baseURL == "" {
		baseURL = "wss://fstream.binance.com/stream" // fallback default
	}

	return fmt.Sprintf("%s?streams=%s", baseURL, streamsURL)
}
