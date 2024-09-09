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
	programLevel := new(slog.LevelVar)
	programLevel.Set(slog.LevelDebug)
	logHandler := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: programLevel})
	slog.SetDefault(slog.New(logHandler))

	httpServer := http_server.NewServer(INIT_WORKER_COUNT, MAXIMUM_WORKER_COUNT)

	log.Fatal(http.ListenAndServe(PORT, httpServer))
}
