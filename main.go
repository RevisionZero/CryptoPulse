package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

func main() {
	// 1. Change to the Market Streams endpoint
	// Combined streams use the "/stream?streams=" path
	// Streams must be lowercase: symbol@streamName
	streams := "btcusdt@bookTicker/ethusdt@bookTicker"
	endpoint := fmt.Sprintf("wss://fstream.binance.com/stream?streams=%s", streams)

	log.Printf("Connecting to %s...", endpoint)

	// 2. Establish the connection
	conn, _, dialErr := websocket.DefaultDialer.Dial(endpoint, nil)
	if dialErr != nil {
		log.Fatal("Dial error:", dialErr)
	}
	defer conn.Close()

	// Handle OS interrupt signals (Ctrl+C)
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	// 3. Start a goroutine to read messages
	// Note: Combined streams wrap data in a {"stream": "name", "data": {...}} object

	type BreakerState int

	const (
		Closed BreakerState = iota
		Open
		HalfOpen
	)

	type CircuitBreaker struct {
		State         BreakerState
		FailThreshold int // Number of fails before circuit breaker trips
		SuccessNeeded int // Number of successes needed for circuit breaker to close
		FailCount     int // Current fail count
		SuccessCount  int // Current success count
	}

	go func() {

		cb := CircuitBreaker{Closed, 9, 20, 0, 0}
		//errorCounter := 0
		for {
			// Closed
			_, message, err := conn.ReadMessage()
			if err != nil {
				// Closed under threshold
				cb.FailCount++
				//errorCounter++
				if cb.FailCount > cb.FailThreshold {
					// Open, threshold crossed
					maxWait := 60000
					baseWait := 1000
					for baseWait > 0 {
						// Sleep for 1 second, then retry connection
						waitTime := rand.Float64() * float64(baseWait)
						time.Sleep(time.Duration(waitTime) * time.Millisecond)
						conn, _, dialErr = websocket.DefaultDialer.Dial(endpoint, nil)
						if dialErr == nil {
							// If reconnecting worked, enter half-open state with limited re-trys
							// Initiate counters for success and fails of re-trys
							//successCounter := 0
							cb.FailCount = 0
							//errorCounter = 0
							// Keep retrying until 20 successes are reached
							for cb.SuccessCount <= cb.SuccessNeeded {
								_, message, err = conn.ReadMessage()
								if err == nil {
									cb.SuccessCount++
								} else {
									cb.FailCount++
								}
								// If while re-retrying, the fails reach 10, enter open state again
								if cb.FailCount > cb.FailThreshold {
									if baseWait < maxWait {
										baseWait *= 2
									}
									cb.FailCount = 0
									break
								}
							}
							// Only enter closed after successful re-trys in half open state
							if cb.SuccessCount >= cb.SuccessNeeded {
								baseWait = -1
								cb.FailCount = 0
							}
						} else {
							if baseWait < maxWait {
								baseWait *= 2
							}
						}
					}
				} else {
					log.Println("Read error:", err)
					conn, _, err = websocket.DefaultDialer.Dial(endpoint, nil)
					//return
				}
			} else {
				// This will now print the live price updates pushed by Binance
				fmt.Printf("Live Stream Data: %s\n", message)
			}
		}
	}()

	log.Println("Connection established. Listening for 10 seconds...")

	// Keep the main thread alive
	select {
	case <-interrupt:
		log.Println("Interrupt received, closing connection...")
		return
	case <-time.After(10 * time.Second):
		log.Println("Test finished.")
	}
}
