package engine

import (
	"main/pkg/models"
	"maps"
	"sync"
	"time"
)

const samplingPeriod = 100 * time.Millisecond

func Sampler(symbols []string, latestPrices map[string]float64, lock *sync.RWMutex, slidingWindows map[string]*models.RingBuffer, sampledDataChan chan<- map[string][]float64) {
	ticker := time.NewTicker(samplingPeriod)
	defer ticker.Stop()

	for range ticker.C {
		lock.RLock()
		if len(latestPrices) == 0 {
			lock.RUnlock()
			continue
		}
		sample := maps.Clone(latestPrices)
		lock.RUnlock()
		sampledData := make(map[string][]float64)
		for _, symbol := range symbols {
			slidingWindows[symbol].Add(sample[symbol])
			sampledData[symbol] = slidingWindows[symbol].GetAll()
		}

		sampledDataChan <- sampledData

	}
}
