package main

import (
	"coin_chase/http_server"
	"log"
	"log/slog"
	"net/http"
	"os"
)

const PORT = ":8888"
const (
	INIT_WORKER_COUNT    = 10
	MAXIMUM_WORKER_COUNT = 10
)

func main() {
	var programLevel = new(slog.LevelVar)
	h := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: programLevel})
	slog.SetDefault(slog.New(h))
	programLevel.Set(slog.LevelDebug)
	// initializeWorkers
	gameServer := http_server.NewServer(INIT_WORKER_COUNT, MAXIMUM_WORKER_COUNT)

	log.Fatal(http.ListenAndServe(PORT, gameServer))
}
