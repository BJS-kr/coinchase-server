package worker_pool

import (
	"coin_chase/game"
	"coin_chase/worker_pool/worker_status"

	"context"
	"log/slog"

	"time"
)

const WORKER_HEALTH_CHECK_TIMEOUT = time.Second * 10

func HealthCheckAndRevive(intervalSec int, maximumWorkerCount int) {
	workerPool := GetWorkerPool()

	for {
		time.Sleep(time.Second * time.Duration(intervalSec))

		for workerId, worker := range workerPool.GetCopiedPool() {
			if worker.GetStatus() == worker_status.TERMINATED {
				slog.Debug("terminated worker discovered")

				workerPool.Delete(workerId)
				workerPool.LaunchWorkers(1, maximumWorkerCount)
			}

			timeout, cancel := context.WithTimeout(context.Background(), WORKER_HEALTH_CHECK_TIMEOUT)
			worker.HealthChecker <- game.Signal

			select {
			case <-timeout.Done():
				{
					workerPool.Delete(workerId)
					// worker와 TCP Receiver가 mutually terminate되므로
					// health check 이상시 worker에게 force exit signal을 전송하면 자원들이 정리된다.
					// 물론 worker가 이미 panic되었을 경우가 가장 많겠지만 혹시 모를 leak을 방지하기 위한 것이다.
					worker.ForceExitSignal <- game.Signal
					close(worker.ForceExitSignal)

					workerPool.LaunchWorkers(1, maximumWorkerCount)
				}
			case <-worker.HealthChecker:
				cancel()
			}
		}

	}
}
