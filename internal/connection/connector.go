package connection

import (
	"log/slog"
	"math/rand/v2"
	"time"

	"github.com/gorilla/websocket"
)

type Connection struct {
	conn     *websocket.Conn
	endpoint string
}

type MessageResponse struct {
	message []byte
	err     error
}

func Connector(symbols []string, dataChan chan<- []byte, closeChan chan bool) {

	slog.Info("Connecting to Binance for symbols: %v", symbols)

	endpoint := constructBinanceEndpoint(symbols)
	conn := Connection{endpoint: endpoint}
	dialErr := conn.dial()
	if dialErr != nil {
		slog.Info("Dial error:", dialErr)
		return
	}
	defer conn.conn.Close()

	cb := CircuitBreaker{Closed, 9, 20, 0, 0, true}

	internalMsgChan := make(chan MessageResponse, 100)

	// Loop to have non-blocking read from connection
	go func() {
		for {
			_, message, err := conn.conn.ReadMessage()
			if err != nil {
				close(internalMsgChan)
				return
			}
			resp := MessageResponse{message, err}
			select {

			case internalMsgChan <- resp:
			case <-closeChan:
				return
			}
		}
	}()

	for {
		select {
		case <-closeChan:
			return
		case msg := <-internalMsgChan:
			if cb.requestPermission() {
				dataChan <- msg.message

				if msg.err != nil {
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

}

func (conn *Connection) dial() (dialErr error) {
	conn.conn, _, dialErr = websocket.DefaultDialer.Dial(conn.endpoint, nil)
	return dialErr
}
