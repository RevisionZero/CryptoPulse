package engine

import (
	"main/pkg/models"
	"sync"
	"time"
)

const samplingPeriod = 100 * time.Millisecond

func Sampler(symbols map[string]*models.SymbolAttributes, lock *sync.RWMutex, sampledDataChan chan<- map[string][]float64) {
	ticker := time.NewTicker(samplingPeriod)
	defer ticker.Stop()

	for range ticker.C {
		lock.RLock()
		// if len(latestPrices) == 0 {
		// 	lock.RUnlock()
		// 	continue
		// }
		sample := make(map[string]float64)
		for symbol, symbolAttr := range symbols {
			sample[symbol] = symbolAttr.LatestPrice
		}
		lock.RUnlock()
		sampledData := make(map[string][]float64)
		for symbol, symbolAttr := range symbols {
			price, ok := sample[symbol]
			if !ok {
				continue
			}
			window := symbolAttr.SlidingWindow
			// if !ok || window == nil {
			// 	continue
			// }
			window.Add(price)
			sampledData[symbol] = window.GetAll()
		}

		sampledDataChan <- sampledData

	}
}
