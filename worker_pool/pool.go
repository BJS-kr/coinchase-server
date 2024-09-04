package worker_pool

import (
	"coin_chase/game"
	"coin_chase/worker_pool/worker_status"
	"errors"
	"log/slog"
	"net"
)

func (w *Worker) SetClientInformation(userId string, clientIP *net.IP, clientPort int) {
	if w.Status != worker_status.PULLED_OUT {
		w.ForceExitSignal <- game.Signal
		slog.Debug("INVALID STATUS CHANGE: WORKER STATUS NOT \"IDLE\"")

		return
	}

	w.Status = worker_status.CLIENT_INFORMATION_RECEIVED
	w.OwnerUserID = userId
	w.ClientIP = clientIP
	w.ClientPort = clientPort
}

func (w *Worker) StartSendUserRelatedDataToClient() bool {
	if w.Status != worker_status.CLIENT_INFORMATION_RECEIVED {
		w.ForceExitSignal <- game.Signal
		slog.Debug("INVALID STATUS CHANGE: WORKER STATUS NOT \"CLIENT_INFORMATION_RECEIVED\"")

		return false
	}

	w.Status = worker_status.WORKING

	go w.CollectedSendUserRelatedDataToClient(w.OwnerUserID, w.ClientIP, w.ClientPort, w.StopClientSendSignal)

	return true
}

func GetWorkerPool() *WorkerPool {
	if !workerPool.Initialized {
		workerPool = WorkerPool{
			Pool:        make(map[string]*Worker),
			Initialized: true,
		}
	}

	return &workerPool
}

func (wp *WorkerPool) GetAvailableWorkerCount() int {
	count := 0

	for _, worker := range wp.Pool {
		if worker.Status == worker_status.IDLE {
			count++
		}
	}

	return count
}

func (wp *WorkerPool) Pull() (*Worker, error) {
	wp.mtx.Lock()

	defer wp.mtx.Unlock()

	for _, worker := range wp.Pool {
		if worker.Status == worker_status.IDLE {
			worker.Status = worker_status.PULLED_OUT

			slog.Info("Worker Pulled Out")

			return worker, nil
		}
	}

	return nil, errors.New("worker currently not available")
}

func (wp *WorkerPool) Put(workerId string, worker *Worker) bool {
	if worker.Status != worker_status.WORKING && worker.Status != worker_status.IDLE {
		worker.ForceExitSignal <- game.Signal
		slog.Debug("INVALID STATUS CHANGE: WORKER STATUS NOT \"WORKING\" OR \"IDLE\"")

		return false
	}

	if worker.Status == worker_status.WORKING {
		worker.StopClientSendSignal <- game.Signal
	}

	worker.Status = worker_status.IDLE
	worker.OwnerUserID = ""
	worker.ClientIP = nil
	worker.ClientPort = 0

	wp.Pool[workerId] = worker

	slog.Info("Put Worker to pool")

	return true
}

func (wp *WorkerPool) PoolSize() int {
	return len(wp.Pool)
}

func (wp *WorkerPool) GetWorkerByUserId(userId string) (string, *Worker, error) {
	for workerId, worker := range wp.Pool {
		if worker.OwnerUserID == userId {
			return workerId, worker, nil
		}
	}

	return "", nil, errors.New("worker not found")
}

func (wp *WorkerPool) MakeWorker(port int) *Worker {
	return &Worker{
		Port:                 port,
		Status:               worker_status.IDLE,
		HealthChecker:        make(chan game.EmptySignal),
		ForceExitSignal:      make(chan game.EmptySignal),
		StopClientSendSignal: make(chan game.EmptySignal),
	}
}

func (wp *WorkerPool) BroadcastgameMapUpdate(gameMapUpdateChannel chan game.EmptySignal) {
	for range gameMapUpdateChannel {
		for _, worker := range wp.Pool {
			if worker.Status == worker_status.WORKING {
				worker.BroadcastUpdateChannel <- game.Signal
			}
		}
	}
}
