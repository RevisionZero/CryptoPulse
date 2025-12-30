package main

import (
	"log"
	"main/internal/connection"
	"main/internal/engine"
	"main/internal/hub"
	"net/http"
	"os"
	"os/signal"
)

func main() {
	symbols := []string{"BTCUSDT", "ETHUSDT"}

	const channelCapacity = 100
	dataChan := make(chan []byte, channelCapacity)
	matrixChan := make(chan map[string]map[string]float64, 1)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	// Create client manager
	clientManager := hub.NewClientManager()

	// Start broadcaster goroutine
	go clientManager.Broadcaster()

	// Start HTTP server
	http.HandleFunc("/ws", clientManager.WSHandler)
	go func() {
		log.Println("WebSocket server starting on :8080")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	// Start connector in goroutine
	go connection.Connector(symbols, dataChan)

	go engine.Synchronizer(symbols, dataChan, matrixChan)

	// Read and broadcast messages
	for {
		select {
		case msg := <-matrixChan:
			clientManager.Broadcast(msg)
		case <-interrupt:
			log.Println("Interrupt received, closing connection...")
			return
		}
	}
}
