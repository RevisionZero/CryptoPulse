package engine

import (
	"main/pkg/models"
	"sync"
)

const slidingWindowSize = 600

func Synchronizer(symbols []string, dataStream <-chan []byte) {
	latestPrices := make(map[string]float64)

	slidingWindows := make(map[string]*models.RingBuffer)

	sampledDataChan := make(chan map[string][]float64, 1)

	for _, symbol := range symbols {
		slidingWindows[symbol] = models.NewRingBuffer(slidingWindowSize)
	}

	var lock sync.RWMutex

	go PriceUpdater(latestPrices, dataStream, &lock)

	go Sampler(symbols, latestPrices, &lock, slidingWindows, sampledDataChan)

	go PCCMatrixCalculator(sampledDataChan, symbols)
}
