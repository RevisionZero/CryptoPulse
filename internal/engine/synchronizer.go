package engine

import (
	"main/pkg/models"
	"sync"
)

const slidingWindowSize = 600

func Synchronizer(symbols map[string]*models.SymbolAttributes, dataStream <-chan []byte, sampleChan chan models.Sample, symbolLock *sync.Mutex) {

	go PriceUpdater(symbols, dataStream, symbolLock)

	go Sampler(symbols, symbolLock, sampleChan)
}
