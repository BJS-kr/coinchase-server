package main

import (
	"coin_chase/game"
	"coin_chase/http_server"
	"coin_chase/worker_pool"
	"fmt"
	"log"
	"net/http"
)

func Initialize(initWorkerCount, maximumWorkerCount int) {
	statusChannel := make(chan *game.Status)
	worker_pool.LaunchWorkers(initWorkerCount, statusChannel, maximumWorkerCount)
	workerPool := worker_pool.GetWorkerPool()

	if workerPool.GetAvailableWorkerCount() != initWorkerCount {
		panic(fmt.Sprintf("worker pool initialization failed. initialized count: %d, expected count: %d", len(workerPool.Pool), initWorkerCount))
	}

	gameMapUpdateChannel := make(chan game.EmptySignal)

	go worker_pool.HealthCheckAndRevive(10, statusChannel, maximumWorkerCount)

	gameMap := game.GetGameMap()

	go gameMap.StartUpdateObjectPosition(statusChannel, gameMapUpdateChannel)
	go workerPool.BroadcastSignal(gameMapUpdateChannel)

	gameMap.InitializeCoins()
	gameMap.InitializeItems()

	go game.SendCoinMoveSignalIntervally(statusChannel, 500)

	httpServer := http_server.NewServer()

	log.Fatal(http.ListenAndServe(PORT, httpServer))
}
