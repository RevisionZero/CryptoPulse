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

type RingBuffer struct {
	data  []float64
	index int
	size  int
}

func NewRingBuffer(size int) *RingBuffer {
	return &RingBuffer{
		data: make([]float64, size),
		size: size,
	}
}

func (rb *RingBuffer) Add(value float64) {
	rb.data[rb.index] = value
	rb.index = (rb.index + 1) % rb.size
}

func (rb *RingBuffer) GetAll() []float64 {
	result := make([]float64, rb.size)
	copy(result, rb.data[rb.index:])
	copy(result[rb.size-rb.index:], rb.data[:rb.index])
	return result
}
