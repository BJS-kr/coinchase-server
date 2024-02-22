package main

import (
	"log"
	"log/slog"
	"multiplayer_server/server"
	"net/http"
	"os"
)

func main() {
	var programLevel = new(slog.LevelVar)
	h := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: programLevel})
	slog.SetDefault(slog.New(h))
	programLevel.Set(slog.LevelDebug)
	// initializeWorkers
	gameServer := server.NewServer()
	log.Fatal(http.ListenAndServe(":8888", gameServer))
}
