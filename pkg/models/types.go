package models

import (
	"main/pkg/utils"
	"time"
)

// CombinedStream represents the outer JSON wrapper from Binance
type CombinedStream struct {
	Stream string        `json:"stream"`
	Data   BinanceTicker `json:"data"`
}

// BinanceTicker represents the specific fields we need for PCC
type BinanceTicker struct {
	EventType string `json:"e"`
	Symbol    string `json:"s"`
	BestBid   string `json:"b"`
	BestAsk   string `json:"a"`
	TransTime int64  `json:"T"`
	EventTime int64  `json:"E"`
}

type SymbolAttributes struct {
	LatestPrice   float64
	SlidingWindow *utils.RingBuffer
	ClientCounter int       // Maintain a count of clients actively using this symbol
	Close         chan bool // Channel to signal close unused websocket for symbol
}

// This struct encapsulates the sampled data sent at each tick as well as the time the tick
// started for latency timing
type Sample struct {
	Data      map[string][]float64
	StartTime time.Time
}

// This struct encapsulates data sent to a client
type Frame struct {
	Payload   []byte    // Payload data to send to client
	StartTime time.Time // Start time, the time when the tick happened triggering the calculation
}
