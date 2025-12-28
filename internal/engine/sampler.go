package engine

import (
	"main/pkg/models"
	"maps"
	"sync"
	"time"
)

const samplingPeriod = 100 * time.Millisecond

func Sampler(symbols []string, latestPrices map[string]float64, lock *sync.RWMutex, slidingWindows map[string]*models.RingBuffer, sampledDataChan chan<- models.PriceMutation) {
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

		// sampledData := make(map[string][]float64)
		// for _, symbol := range symbols {
		// 	slidingWindows[symbol].Add(sample[symbol])
		// 	sampledData[symbol] = slidingWindows[symbol].GetAll()
		// }

		// sampledDataChan <- sampledData

		newPrices := make(map[string]float64, len(symbols))
		oldPrices := make(map[string]float64, len(symbols))

		for _, symbol := range symbols {
			price := sample[symbol]
			// Add now returns the "Victim" falling off the 60s window
			oldPrice := slidingWindows[symbol].Add(price)

			newPrices[symbol] = price
			oldPrices[symbol] = oldPrice
		}

		// 3. Send Mutation: Only 2 floats per symbol are sent
		select {
		case sampledDataChan <- models.PriceMutation{
			NewPrices: newPrices,
			OldPrices: oldPrices,
		}:
		default:
			// Non-blocking send: Drop if the calculator is falling behind
		}

	}
}
