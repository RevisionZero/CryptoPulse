package utils

type RingBuffer struct {
	data  []float64
	index int
	size  int
	// Once rb is full, aka fillcounter = size,
	// we know the rb is full.
	fillCounter int
	full        bool
}

func NewRingBuffer(size int) *RingBuffer {
	return &RingBuffer{
		data:        make([]float64, size),
		size:        size,
		fillCounter: 0,
		full:        false,
	}
}

func (rb *RingBuffer) Add(value float64) float64 {
	if rb.fillCounter < rb.size {
		rb.fillCounter++
		if rb.fillCounter == rb.size {
			rb.full = true
		}
	}
	oldValue := rb.data[rb.index]
	rb.data[rb.index] = value
	rb.index = (rb.index + 1) % rb.size
	return oldValue
}

func (rb *RingBuffer) GetAll() []float64 {
	result := make([]float64, rb.size)
	copy(result, rb.data[rb.index:])
	copy(result[rb.size-rb.index:], rb.data[:rb.index])
	return result
}

func (rb *RingBuffer) IsFull() bool {
	return rb.full
}
