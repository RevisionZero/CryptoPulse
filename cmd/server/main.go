package main

import (
	"log/slog"
	"main/internal/hub"
	"net/http"
	"os"
	"os/signal"

	"github.com/joho/godotenv"
)

// main configures logging, starts the websocket backend, and blocks until shutdown.
func main() {

	// Emit structured JSON logs so backend events can be consumed by observability tooling.
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	slog.SetDefault(logger)

	// Load .env file
	if err := godotenv.Load(); err != nil {
		slog.Info("No .env file found, using defaults")
	}

	// Buffer one signal so the process can begin graceful shutdown work immediately.
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	// Buffered channel smooths bursty publish traffic between engine and hub fan-out.
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
