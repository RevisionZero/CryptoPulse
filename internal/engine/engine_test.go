package engine

import (
	"strconv"
	"testing"
	"time"
)

func FeedKnownSignal(dataStream chan<- []byte) {
	for i := 1.0; i <= 600.0; i++ {
		btcJson := []byte(`{"stream":"btcusdt@bookTicker","data":{"s":"BTCUSDT","b":"` + strconv.FormatFloat(60000+i, 'f', 2, 64) + `","a":"` + strconv.FormatFloat(60001+i, 'f', 2, 64) + `"}}`)
		ethJson := []byte(`{"stream":"ethusdt@bookTicker","data":{"s":"ETHUSDT","b":"` + strconv.FormatFloat(3000+i, 'f', 2, 64) + `","a":"` + strconv.FormatFloat(3001+i, 'f', 2, 64) + `"}}`)

		dataStream <- btcJson
		dataStream <- ethJson

		time.Sleep(10 * time.Millisecond)
	}
}

func Test_Engine(t *testing.T) {
	// 1. Setup the same channels as your real app
	dataStream := make(chan []byte, 100)
	symbols := []string{"BTCUSDT", "ETHUSDT"}

	// 2. Start your existing engine components
	go Synchronizer(symbols, dataStream)

	// 3. Instead of connecting to Binance, feed your known signal
	FeedKnownSignal(dataStream)
}
