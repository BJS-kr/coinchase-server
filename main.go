package main

import (
	"log/slog"
	"multiplayer_server/server"
	"os"
)

func main() {
	var programLevel = new(slog.LevelVar)
	h := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: programLevel})
	slog.SetDefault(slog.New(h))
	programLevel.Set(slog.LevelDebug)
	// initializeWorkers
	server.RunServer()
}
