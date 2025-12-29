package models

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
