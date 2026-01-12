package engine

import (
	"main/pkg/models"
	"sync"
	"time"
)

const samplingPeriod = 100 * time.Millisecond

func Sampler(symbols map[string]*models.SymbolAttributes, symbolLock *sync.Mutex, sampledDataChan chan<- map[string][]float64) {
	ticker := time.NewTicker(samplingPeriod)
	defer ticker.Stop()

	// sample := make(map[string]float64)
	sampledData := make(map[string][]float64)

	for range ticker.C {

		symbolLock.Lock()
		for symbol, symbolAttr := range symbols {
			symbolAttr.SlidingWindow.Add(symbolAttr.LatestPrice)
			sampledData[symbol] = symbolAttr.SlidingWindow.GetAll()
		}
		symbolLock.Unlock()

		sampledDataChan <- sampledData

	}
}
