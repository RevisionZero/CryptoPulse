package connection

import (
	"fmt"
	"strings"
)

func constructBinanceEndpoint(symbols []string) string {
	streamsURL := ""
	for _, symbol := range symbols {
		streamsURL += fmt.Sprintf("%s@bookTicker/", strings.ToLower(symbol))
	}
	streamsURL = streamsURL[:len(streamsURL)-1]
	return fmt.Sprintf("wss://fstream.binance.com/stream?streams=%s", streamsURL)
}
