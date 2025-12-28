package utils

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
