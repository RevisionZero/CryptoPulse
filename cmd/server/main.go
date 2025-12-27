package main

import (
	"log"
	"main/internal/connection"
	"main/internal/engine"
	"os"
	"os/signal"
	"time"
)

// func main1() {
// 	// STEP 1: Create the Channels (The Pipes)
// 	// rawDataChan: Ingest -> Calc
// 	rawDataChan := make(chan []byte, 100)

// 	// resultChan: Calc -> Hub
// 	resultChan := make(chan models.CorrelationResult, 10)

// 	// STEP 2: Start the Hub (Consumer of Results)
// 	// The Hub sits waiting to broadcast to UI
// 	hub := hub.NewHub(resultChan)
// 	go hub.Run()

// 	// STEP 3: Start the Engine (Consumer of Raw Data / Producer of Results)
// 	// We pass the "input" pipe and the "output" pipe
// 	engine := calc.NewEngine(rawDataChan, resultChan)
// 	go engine.Start()

// 	// STEP 4: Start the Connector (The Producer)
// 	// We pass the "output" pipe where it should dump Binance data
// 	endpoint := "wss://fstream.binance.com/stream?streams=btcusdt@bookTicker"
// 	go connection.Connector(endpoint, rawDataChan)

// 	// Block main so the program doesn't exit
// 	select {}
// }

func main() {
	symbols := []string{"btcusdt", "ethusdt"}
	dataChan := make(chan []byte, 100)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	// Start connector in goroutine
	go connection.Connector(symbols, dataChan)

	go engine.Synchronizer(symbols, dataChan)

	// Read and print messages
	for {
		select {
		// case msg := <-dataChan:
		// 	fmt.Printf("Received: %s\n", string(msg))
		case <-interrupt:
			log.Println("Interrupt received, closing connection...")
			return
		case <-time.After(10 * time.Second):
			log.Println("Test finished.")
			return
		}
	}
}
