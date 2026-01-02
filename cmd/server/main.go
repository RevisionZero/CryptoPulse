package main

import (
	"log"
	"main/internal/hub"
	"net/http"
	"os"
	"os/signal"
)

func main() {

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	broadcast := make(chan map[string][]float64, 256)

	hub := hub.NewHub(broadcast)

	http.HandleFunc("/ws", hub.WSHandler)
	go func() {
		log.Println("WebSocket server starting on :8080")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()

	go hub.Run()

	for {
		select {
		case <-interrupt:
			log.Println("Interrupt received, closing connection...")
			return
		}
	}
}
