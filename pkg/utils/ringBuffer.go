package utils

// RingBuffer keeps a fixed-size rolling sequence of float64 values.
type RingBuffer struct {
	data  []float64
	index int
	size  int
}

// NewRingBuffer allocates a fixed-size circular buffer.
func NewRingBuffer(size int) *RingBuffer {
	return &RingBuffer{
		data: make([]float64, size),
		size: size,
	}
}

// Add writes a value at the current cursor and advances to the next slot.
func (rb *RingBuffer) Add(value float64) {
	rb.data[rb.index] = value
	rb.index = (rb.index + 1) % rb.size
}

// GetAll returns the buffer ordered from oldest to newest value.
func (rb *RingBuffer) GetAll() []float64 {
	result := make([]float64, rb.size)
	copy(result, rb.data[rb.index:])
	copy(result[rb.size-rb.index:], rb.data[:rb.index])
	return result
}
