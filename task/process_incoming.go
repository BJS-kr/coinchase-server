package task

import (
	"log/slog"
	"multiplayer_server/game_map"
	"multiplayer_server/protodef"
	"multiplayer_server/worker_pool"
	"sync"
)

func ProcessIncoming(worker *worker_pool.Worker, initWorker *sync.WaitGroup, statusReceiver <-chan *protodef.Status, workerPool *worker_pool.WorkerPool, mutualTerminationSignal chan bool) {
	defer SendMutualTerminationSignal(mutualTerminationSignal)

	initWorker.Done()
	slog.Info("Worker Initialized")

	for {
		select {
		case status := <-statusReceiver:
			items := make([]game_map.Item, 0)

			for _, item := range status.Items {
				items = append(items, game_map.Item{
					Id:     item.Id,
					Name:   item.Name,
					Amount: item.Amount,
				})
			}

			safeStatus := game_map.Status{
				Id: status.Id,
				CurrentPosition: game_map.Position{
					X: status.CurrentPosition.X,
					Y: status.CurrentPosition.Y,
				},
				Items: items,
			}

			game_map.GameMap.UpdateUserPosition(&safeStatus)
	

		case <-worker.ForceExitSignal:
			// panic하는 이유는 mutual termination을 실행해야하기 때문이다.
			// 이에 따라 자동으로 UDP Receiver도 종료될 것이다.
			panic("forced exit occurred by signal")

		case <-worker.HealthChecker:
			worker.HealthChecker <- true

		case <-mutualTerminationSignal:
			// 이 시그널이 도착했다는 것은 UDP receiver가 이미 panic했다는 의미이다. 그러므로 단순히 return하면 된다.
			return
		}
	}
}
