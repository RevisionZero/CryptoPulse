package engine

import (
	"main/pkg/models"
	"sync"
	"time"
)

const samplingPeriod = 100 * time.Millisecond

func Sampler(symbols map[string]*models.SymbolAttributes, symbolLock *sync.Mutex, sampledDataChan chan<- models.Sample) {
	ticker := time.NewTicker(samplingPeriod)
	defer ticker.Stop()

	// sample := make(map[string]float64)
	sampledData := make(map[string][]float64)

	for tickTime := range ticker.C {

		symbolLock.Lock()
		for symbol, symbolAttr := range symbols {
			symbolAttr.SlidingWindow.Add(symbolAttr.LatestPrice)
			sampledData[symbol] = symbolAttr.SlidingWindow.GetAll()
		}
		symbolLock.Unlock()
		sampledDataChan <- models.Sample{Data: sampledData, StartTime: tickTime}

	}
}
