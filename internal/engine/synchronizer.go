package engine

import (
	"main/pkg/models"
	"sync"
)

const slidingWindowSize = 600

func Synchronizer(symbols map[string]*models.SymbolAttributes, dataStream <-chan []byte, sampledDataChan chan map[string][]float64) {
	// latestPrices := make(map[string]float64)

	// slidingWindows := make(map[string]*utils.RingBuffer)

	// for _, symbol := range symbols {
	// 	slidingWindows[symbol] = utils.NewRingBuffer(slidingWindowSize)
	// }

	var lock sync.RWMutex

	go PriceUpdater(symbols, dataStream, &lock)

	go Sampler(symbols, &lock, sampledDataChan)
}
