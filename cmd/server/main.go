package main

import (
	"log/slog"
	"main/internal/hub"
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

	// Load .env file
	if err := godotenv.Load(); err != nil {
		slog.Info("No .env file found, using defaults")
	}

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	broadcast := make(chan map[string][]float64, 256)

	hub := hub.NewHub(broadcast)

	http.HandleFunc("/ws", hub.WSHandler)
	go func() {
		slog.Info("WebSocket server starting on :8080")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			slog.Info("HTTP server error: %v", err)
			os.Exit(1)
		}
	}()

	go hub.Run()

	for {
		select {
		case <-interrupt:
			slog.Info("Interrupt received, closing connection...")
			return
		}
	}
}
