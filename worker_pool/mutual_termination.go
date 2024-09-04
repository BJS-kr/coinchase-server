package worker_pool

import (
	"coin_chase/worker_pool/worker_status"
	"context"
	"log/slog"
)

func CollectWorkerForMutualTermination(worker *Worker, mutualCancel context.CancelFunc) func() {
	return func() {
		if r := recover(); r != nil {
			// close는 이 채널을 구독하는 모든 goroutine에게 zero value를 보내므로 단순히 close하는 것만으로도 모든 goroutine을 정리할 수 있다.
			mutualCancel()
			worker.Status = worker_status.TERMINATED
			slog.Debug("mutual termination signal sent")
		}
	}
}
