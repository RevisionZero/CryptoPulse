package main

import (
	"log"
	"main/internal/hub"
	"net/http"
	"os"
	"os/signal"
)

func main() {

	// const channelCapacity = 100
	// dataChan := make(chan []byte, channelCapacity)
	// matrixChan := make(chan map[string]map[string]float64, 1)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	broadcast := make(chan map[string][]float64, 256)

	// Create Hub
	hub := hub.NewHub(broadcast)

	// Start broadcaster goroutine
	// go hub.Broadcaster()

	// Start HTTP server
	http.HandleFunc("/ws", hub.WSHandler)
	go func() {
		log.Println("WebSocket server starting on :8080")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	go hub.Run()

	// Read and broadcast messages
	for {
		select {
		case <-interrupt:
			log.Println("Interrupt received, closing connection...")
			return
		}
	}
}
