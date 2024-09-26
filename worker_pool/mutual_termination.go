package worker_pool

import (
	"coin_chase/worker_pool/worker_status"
	"context"
	"log/slog"
)

func CollectWorkerForMutualTermination(worker *Worker, mutualCancel context.CancelFunc) func() {
	return func() {
		if recover() != nil {
			slog.Error("panic occurred in mutual termination")
		}
		mutualCancel()
		worker.ChangeStatus(worker_status.TERMINATED)
		slog.Debug("mutual termination signal sent")
	}
}
