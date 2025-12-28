package engine

import (
	"main/pkg/models"
	"sync"
)

const slidingWindowSize = 600

func Synchronizer(symbols []string, dataStream <-chan []byte) {
	latestPrices := make(map[string]float64)

	slidingWindows := make(map[string]*models.RingBuffer)

	const channelCapacity = 2
	sampledDataChan := make(chan models.PriceMutation, channelCapacity)

	for _, symbol := range symbols {
		slidingWindows[symbol] = models.NewRingBuffer(slidingWindowSize)
	}

	var lock sync.RWMutex

	go PriceUpdater(latestPrices, dataStream, &lock)

	go Sampler(symbols, latestPrices, &lock, slidingWindows, sampledDataChan)

	go PCCMatrixCalculator(sampledDataChan, symbols)
}
