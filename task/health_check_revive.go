package task

import (
	"context"
	"log/slog"
	"multiplayer_server/worker_pool"

	"time"
)

func HealthCheckAndRevive(intervalSec int) {
	workerPool := worker_pool.GetWorkerPool()

	for {
		time.Sleep(time.Second * time.Duration(intervalSec))

		for workerID, worker := range workerPool.Pool {
			if worker.Status == worker_pool.TERMINATED {
				slog.Debug("terminated worker discovered")
				delete(workerPool.Pool, workerID)
				LaunchWorkers(1)
			}

			timeout, cancel := context.WithTimeout(context.Background(), time.Second*10)
			worker.HealthChecker <- true

			select {
			case <-timeout.Done():
				{
					delete(workerPool.Pool, workerID)
					// worker와 UDP Receiver가 mutually terminated되므로
					// health check 이상시 worker에게 force exit signal을 전송하면 자원들이 정리된다.
					// 물론 worker가 이미 panic되었을 경우가 가장 많겠지만 혹시 모를 leak을 방지하기 위한 것이다.
					worker.ForceExitSignal <- true
					close(worker.ForceExitSignal)

					LaunchWorkers(1)
				}
			case <-worker.HealthChecker:
				cancel()
			}
		}
	}
}
