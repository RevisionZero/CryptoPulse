package engine

import (
	"encoding/json"
	"log"
	"main/pkg/models"
	"strconv"
	"time"
)

func Synchronizer(symbols []string, dataStream <-chan []byte) {
	var latestPrices map[string]chan float64
	for _, symbol := range symbols {
		latestPrices[symbol] = make(chan float64)
	}

	// Start the sampler
	go PriceUpdater(latestPrices, dataStream)

	go Sampler(symbols, latestPrices)
}

func PriceUpdater(latestPrices map[string]chan float64, dataStream <-chan []byte) {

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

		latestPrices[envelope.Data.Symbol] <- (bid + ask) / 2

	}
}

func Sampler(symbols []string, latestPrices map[string]chan float64) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for range ticker.C {
		// Sampling logic here
	}
}
