package main

import (
	"log"
	"main/internal/connection"
	"main/internal/engine"
	"os"
	"os/signal"
)

func main() {
	symbols := []string{"BTCUSDT", "ETHUSDT"}

	const channelCapacity = 100
	dataChan := make(chan []byte, channelCapacity)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	// Start connector in goroutine
	go connection.Connector(symbols, dataChan)

	go engine.Synchronizer(symbols, dataChan)

	// Wait for interrupt signal
	for {
		select {
		case <-interrupt:
			log.Println("Interrupt received, closing connection...")
			return
		}
	}
}
