package engine

import (
	"encoding/json"
	"log"
	"main/pkg/models"
	"maps"
	"strconv"
	"sync"
	"time"
)

func Synchronizer(symbols []string, dataStream <-chan []byte) {
	latestPrices := make(map[string]float64)

	const slidingWindowSize = 600

	slidingWindows := make(map[string]*models.RingBuffer)

	sampledDataChan := make(chan map[string][]float64, 100)

	for _, symbol := range symbols {
		slidingWindows[symbol] = models.NewRingBuffer(slidingWindowSize)
	}

	var lock sync.RWMutex

	go PriceUpdater(latestPrices, dataStream, &lock)

	go Sampler(symbols, latestPrices, &lock, slidingWindows, sampledDataChan)
}

func PriceUpdater(latestPrices map[string]float64, dataStream <-chan []byte, lock *sync.RWMutex) {

	// var latestValues map[string]chan float64
	// for _, symbol := range symbols {
	// 	latestValues[symbol] = make(chan float64)
	// }
	for {
		rawData := <-dataStream
		var envelope models.CombinedStream

		// 1. Unmarshal JSON into the struct
		if err := json.Unmarshal(rawData, &envelope); err != nil {
			log.Printf("Error parsing JSON: %v", err)
			log.Printf("Raw JSON: %s", string(rawData))
			continue
		}

		// log.Printf("Symbol: %s, Best Bid: %s, Best Ask: %s, Transaction Time: %d, Event Time: %d\n\n\n",
		// 	envelope.Data.Symbol,
		// 	envelope.Data.BestBid,
		// 	envelope.Data.BestAsk,
		// 	envelope.Data.TransTime,
		// 	envelope.Data.EventTime)

		bid, err := strconv.ParseFloat(envelope.Data.BestBid, 64)
		if err != nil {
			log.Printf("Error parsing bid: %v", err)
			continue
		}

		ask, err := strconv.ParseFloat(envelope.Data.BestAsk, 64)
		if err != nil {
			log.Printf("Error parsing ask: %v", err)
			continue
		}

		lock.Lock()
		latestPrices[envelope.Data.Symbol] = (bid + ask) / 2
		lock.Unlock()

	}
}

func Sampler(symbols []string, latestPrices map[string]float64, lock *sync.RWMutex, slidingWindows map[string]*models.RingBuffer, sampledDataChan chan<- map[string][]float64) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		lock.RLock()
		sample := maps.Clone(latestPrices)
		sampledData := make(map[string][]float64)
		for _, symbol := range symbols {
			slidingWindows[symbol].Add(sample[symbol])
			sampledData[symbol] = slidingWindows[symbol].GetAll()
		}
		sampledDataChan <- sampledData

		lock.RUnlock()

	}
}
