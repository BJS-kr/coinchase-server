package bootstrap

import (
	"coin_chase/game"
	"coin_chase/worker_pool"
	"fmt"
)

func Run(initWorkerCount, maximumWorkerCount int) {
	statusChannel := make(chan *game.Status)
	workerPool := worker_pool.GetWorkerPool()
	workerPool.LaunchWorkers(initWorkerCount, statusChannel, maximumWorkerCount)

	if workerPool.GetAvailableWorkerCount() != initWorkerCount {
		panic(fmt.Sprintf("worker pool initialization failed. initialized count: %d, expected count: %d", len(workerPool.Pool), initWorkerCount))
	}

	gameMapUpdateChannel := make(chan game.EmptySignal)
	gameMap := game.GetGameMap()

	gameMap.InitializeCoins()
	gameMap.InitializeItems()

	go worker_pool.HealthCheckAndRevive(10, statusChannel, maximumWorkerCount)
	go gameMap.StartUpdateObjectPosition(statusChannel, gameMapUpdateChannel)
	go workerPool.BroadcastSignal(gameMapUpdateChannel)
	go game.SendCoinMoveSignalIntervally(statusChannel, 500)
}
