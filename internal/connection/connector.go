package connection

import (
	"log"
	"os"
	"os/signal"

	"github.com/gorilla/websocket"
)

type Connection struct {
	conn     *websocket.Conn
	endpoint string
}

func connector(endpoint string) {

	log.Printf("Connecting to %s...", endpoint)

	conn := Connection{new(websocket.Conn), endpoint}
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
		cb.waitForPermission(conn.conn)

		// if conn == nil {
		// 	conn, _, dialErr = dial(endpoint)
		// 	if dialErr != nil {
		// 		cb.incrementFails()
		// 		continue
		// 	}
		// }

		_, resp.message, resp.err = conn.ReadMessage()

		if resp.err != nil {
			cb.incrementFails()
			// conn.Close() // Clean up the dead socket
			// conn = nil   // Trigger a re-dial on next loop
			// continue
		}

		// dataChan <- message
	}

}

func (conn *Connection) dial() (dialErr error) {
	conn.conn, _, dialErr = websocket.DefaultDialer.Dial(conn.endpoint, nil)
	return dialErr
}
