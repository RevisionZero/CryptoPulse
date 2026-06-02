package models

import "main/pkg/utils"

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

// Instead of sending entire sliding window with each update, send only the new point,
// the evicted point, and if the ring buffer is full
type DataUpdate struct {
	New     float64
	Evicted float64
	Full    bool
}

// Track running means, centered sums, and centered cross points
type WelfordData struct {
	N     int     // Number of samples currently in the window │
	MeanX float64 // Running mean of X
	MeanY float64 // Running mean of Y
	Mxx   float64 // Σ(x − x̄)² — centered sum of squares for X
	Myy   float64 // Σ(y − ȳ)² — centered sum of squares for Y
	Mxy   float64 // Σ(x − x̄)(y − ȳ) — centered cross-product
}
