package engine

import (
	"encoding/json"
	"log"
	"main/pkg/models"
	"strconv"
	"sync"
)

func PriceUpdater(symbols map[string]*models.SymbolAttributes, dataStream <-chan []byte, lock *sync.RWMutex) {

	for {
		rawData := <-dataStream
		var envelope models.CombinedStream

		// 1. Unmarshal JSON into the struct
		if err := json.Unmarshal(rawData, &envelope); err != nil {
			log.Printf("Error parsing JSON: %v", err)
			log.Printf("Raw JSON: %s", string(rawData))
			continue
		}

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
		symbols[envelope.Data.Symbol].LatestPrice = (bid + ask) / 2
		lock.Unlock()

	}
}
