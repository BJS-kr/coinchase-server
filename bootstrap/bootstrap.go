package bootstrap

import (
	"coin_chase/game"
	"coin_chase/worker_pool"
	"fmt"
)

func Run(initWorkerCount, maximumWorkerCount int) {
	workerPool := worker_pool.GetWorkerPool()
	workerPool.LaunchWorkers(initWorkerCount, maximumWorkerCount)

	if workerPool.GetAvailableWorkerCount() != initWorkerCount {
		panic(fmt.Sprintf("worker pool initialization failed. initialized count: %d, expected count: %d", len(workerPool.Pool), initWorkerCount))
	}

	gameMapUpdateChannel := make(chan game.EmptySignal)
	gameMap := game.GetGameMap()

	gameMap.InitializeCoins()
	gameMap.InitializeItems()

	go worker_pool.HealthCheckAndRevive(10, maximumWorkerCount)
	go gameMap.StartUpdateMap(gameMapUpdateChannel)
	go workerPool.BroadcastSignal(gameMapUpdateChannel)
	go game.SendCoinMoveSignalIntervally(500)
}
