package engine

import (
	"main/pkg/models"
	"sync"
)

const slidingWindowSize = 600

func Synchronizer(symbols map[string]*models.SymbolAttributes, dataStream <-chan []byte, sampledDataChan chan map[string][]float64) {

	var lock sync.RWMutex

	go PriceUpdater(symbols, dataStream, &lock)

	go Sampler(symbols, &lock, sampledDataChan)
}
