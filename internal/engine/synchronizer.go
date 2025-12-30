package engine

import (
	"main/pkg/utils"
	"sync"
)

const slidingWindowSize = 600

func Synchronizer(symbols []string, dataStream <-chan []byte, matrixChan chan<- map[string]map[string]float64) {
	latestPrices := make(map[string]float64)

	slidingWindows := make(map[string]*utils.RingBuffer)

	const channelCapacity = 2
	sampledDataChan := make(chan map[string][]float64, channelCapacity)

	for _, symbol := range symbols {
		slidingWindows[symbol] = utils.NewRingBuffer(slidingWindowSize)
	}

	var lock sync.RWMutex

	go PriceUpdater(latestPrices, dataStream, &lock)

	go Sampler(symbols, latestPrices, &lock, slidingWindows, sampledDataChan)

	go PCCMatrixCalculator(sampledDataChan, symbols, matrixChan)
}
