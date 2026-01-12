package engine

import (
	"main/pkg/models"
	"sync"
)

const slidingWindowSize = 600

func Synchronizer(symbols map[string]*models.SymbolAttributes, dataStream <-chan []byte, sampledDataChan chan map[string][]float64, symbolLock *sync.Mutex) {

	go PriceUpdater(symbols, dataStream, symbolLock)

	go Sampler(symbols, symbolLock, sampledDataChan)
}
