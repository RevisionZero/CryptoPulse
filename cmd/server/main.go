package main

import (
	"encoding/json"
	"log/slog"
	"main/internal/hub"
	"main/pkg/tracker"
	"net/http"
	"os"
	"os/signal"

	"github.com/joho/godotenv"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	if err := godotenv.Load(); err != nil {
		slog.Info("No .env file found, using defaults")
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	dataFilePath := os.Getenv("TRACKER_DATA_FILE")
	if dataFilePath == "" {
		dataFilePath = "./data/user_counts.json"
	}
	t := tracker.New(dataFilePath)

	broadcast := make(chan map[string][]float64, 256)

	h := hub.NewHub(broadcast, t)

	statsToken := os.Getenv("STATS_TOKEN")
	http.HandleFunc("/ws", h.WSHandler)
	http.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
		statsHandler(w, r, t, statsToken)
	})

	go func() {
		slog.Info("WebSocket server starting on :8080")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			slog.Info("HTTP server error", "err", err)
			os.Exit(1)
		}
	}()

	go h.Run()

	for {
		select {
		case <-interrupt:
			slog.Info("Interrupt received, closing connection...")
			return
		}
	}
}

func statsHandler(w http.ResponseWriter, r *http.Request, t *tracker.Tracker, token string) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if token != "" {
		if r.Header.Get("Authorization") != "Bearer "+token {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(t.Counts()); err != nil {
		slog.Info("stats: JSON encode error", "err", err)
	}
}
