package engine

import (
	"encoding/json"
	"log/slog"
	"main/pkg/models"
	"strconv"
	"sync"
)

func PriceUpdater(symbols map[string]*models.SymbolAttributes, dataStream <-chan []byte, symbolLock *sync.Mutex) {

	for {
		rawData := <-dataStream
		var envelope models.CombinedStream

		// 1. Unmarshal JSON into the struct
		if err := json.Unmarshal(rawData, &envelope); err != nil {
			slog.Info("Error parsing JSON: %v", err)
			slog.Info("Raw JSON: %s", string(rawData))
			continue
		}

		bid, err := strconv.ParseFloat(envelope.Data.BestBid, 64)
		if err != nil {
			slog.Info("Error parsing bid: %v", err)
			continue
		}

		ask, err := strconv.ParseFloat(envelope.Data.BestAsk, 64)
		if err != nil {
			slog.Info("Error parsing ask: %v", err)
			continue
		}

		symbolLock.Lock()
		if sym, ok := symbols[envelope.Data.Symbol]; ok && sym != nil {
			sym.LatestPrice = (bid + ask) / 2
		} else {
			slog.Info("Symbol not found or nil in symbols map: %s", envelope.Data.Symbol)
		}
		symbolLock.Unlock()

	}
}
