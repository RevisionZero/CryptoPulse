package connection

import (
	"log"
	"math/rand/v2"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
)

type Connection struct {
	conn     *websocket.Conn
	endpoint string
}

func connector(endpoint string, dataChan chan<- []byte) {

	log.Printf("Connecting to %s...", endpoint)

	conn := Connection{endpoint: endpoint}
	dialErr := conn.dial()
	if dialErr != nil {
		log.Fatal("Dial error:", dialErr)
	}
	defer conn.conn.Close()

	// Handle OS interrupt signals (Ctrl+C)
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	type MessageResponse struct {
		message []byte
		err     error
	}

	cb := CircuitBreaker{Closed, 9, 20, 0, 0, true}
	resp := MessageResponse{[]byte{}, nil}
	for {

		if cb.requestPermission() {
			_, resp.message, resp.err = conn.conn.ReadMessage()

			dataChan <- resp.message

			if resp.err != nil {
				cb.incrementFails()
			} else {
				cb.incrementSuccesses()
			}
		} else {
			maxWait := 60000
			baseWait := 1000
			for baseWait > 0 {
				waitTime := rand.Float64() * float64(baseWait)
				time.Sleep(time.Duration(waitTime) * time.Millisecond)
				dialErr = conn.dial()
				cb.setDialState(dialErr)
				if cb.requestPermission() {
					break
				}
				if baseWait < maxWait {
					baseWait *= 2
				}
			}
		}

	}

}

func (conn *Connection) dial() (dialErr error) {
	conn.conn, _, dialErr = websocket.DefaultDialer.Dial(conn.endpoint, nil)
	return dialErr
}
