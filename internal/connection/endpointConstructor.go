package connection

import "fmt"

func constructBinanceEndpoint(symbols []string) string {
	streamsURL := ""
	for _, symbol := range symbols {
		streamsURL += fmt.Sprintf("%s@bookTicker/", symbol)
	}
	streamsURL = streamsURL[:len(streamsURL)-1]
	return fmt.Sprintf("wss://fstream.binance.com/stream?streams=%s", streamsURL)
}
