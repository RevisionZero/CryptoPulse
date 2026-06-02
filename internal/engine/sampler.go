package engine

import (
	"main/pkg/models"
	"sync"
	"time"
)

const samplingPeriod = 100 * time.Millisecond

func Sampler(symbols map[string]*models.SymbolAttributes, symbolLock *sync.Mutex, sampledDataChan chan<- map[string]models.DataUpdate) {
	ticker := time.NewTicker(samplingPeriod)
	defer ticker.Stop()

	// sample := make(map[string]float64)
	// sampledData := make(map[string][]float64)
	sampledData := make(map[string]models.DataUpdate)

	for range ticker.C {

		symbolLock.Lock()
		for symbol, symbolAttr := range symbols {
			latestPrice := symbolAttr.LatestPrice
			// update.evicted = symbolAttr.SlidingWindow.Add(symbolAttr.LatestPrice)
			// sampledData[symbol] = symbolAttr.SlidingWindow.GetAll()
			sampledData[symbol] = models.DataUpdate{New: latestPrice, Evicted: symbolAttr.SlidingWindow.Add(latestPrice), Full: symbolAttr.SlidingWindow.IsFull()}
		}
		symbolLock.Unlock()

		sampledDataChan <- sampledData

	}
}
