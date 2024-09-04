package worker_pool

import (
	"errors"
	"log/slog"
	"multiplayer_server/global"
	"multiplayer_server/protodef"
	"net"

	"sync"
)

type WorkerStatus int

const WORKER_COUNT = 10
const (
	IDLE = iota + 1
	PULLED_OUT
	CLIENT_INFORMATION_RECEIVED
	WORKING
	TERMINATED
)

type WorkerPool struct {
	mtx         sync.Mutex
	Pool        map[string]*Worker
	Initialized bool
}

var workerPool WorkerPool

type Worker struct {
	ClientIP                             *net.IP
	ClientPort                           int
	Port                                 int
	OwnerUserID                          string
	Status                               WorkerStatus
	StatusReceiver                       <-chan *protodef.Status
	HealthChecker                        chan global.EmptySignal
	ForceExitSignal                      chan global.EmptySignal
	StopClientSendSignal                 chan global.EmptySignal
	CollectedSendUserRelatedDataToClient func(clientID string, clientIP *net.IP, clientPort int, stopClientSendSignal chan global.EmptySignal)
	BroadcastUpdateChannel               chan global.EmptySignal
}

func (w *Worker) SetClientInformation(userId string, clientIP *net.IP, clientPort int) {
	if w.Status != PULLED_OUT {
		w.ForceExitSignal <- global.Signal
		slog.Debug("INVALID STATUS CHANGE: WORKER STATUS NOT \"IDLE\"")

		return
	}

	w.Status = CLIENT_INFORMATION_RECEIVED
	w.OwnerUserID = userId
	w.ClientIP = clientIP
	w.ClientPort = clientPort
}

func (w *Worker) StartSendUserRelatedDataToClient() bool {
	if w.Status != CLIENT_INFORMATION_RECEIVED {
		w.ForceExitSignal <- global.Signal
		slog.Debug("INVALID STATUS CHANGE: WORKER STATUS NOT \"CLIENT_INFORMATION_RECEIVED\"")

		return false
	}

	w.Status = WORKING

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
		if worker.Status == IDLE {
			count++
		}
	}

	return count
}

func (wp *WorkerPool) Pull() (*Worker, error) {
	wp.mtx.Lock()

	defer wp.mtx.Unlock()

	for _, worker := range wp.Pool {
		if worker.Status == IDLE {
			worker.Status = PULLED_OUT

			slog.Info("Worker Pulled Out")

			return worker, nil
		}
	}

	return nil, errors.New("worker currently not available")
}

func (wp *WorkerPool) Put(workerId string, worker *Worker) bool {
	if worker.Status != WORKING && worker.Status != IDLE {
		worker.ForceExitSignal <- global.Signal
		slog.Debug("INVALID STATUS CHANGE: WORKER STATUS NOT \"WORKING\" OR \"IDLE\"")

		return false
	}

	if worker.Status == WORKING {
		worker.StopClientSendSignal <- global.Signal
	}

	worker.Status = IDLE
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
		Status:               IDLE,
		HealthChecker:        make(chan global.EmptySignal),
		ForceExitSignal:      make(chan global.EmptySignal),
		StopClientSendSignal: make(chan global.EmptySignal),
		//UserID와 ClientIP와 ClientPort는 추후 워커가 유저에게 할당 될 때 설정된다.
	}
}

func (wp *WorkerPool) BroadcastGlobalMapUpdate(globalMapUpdateChannel chan global.EmptySignal) {
	for range globalMapUpdateChannel {
		for _, worker := range wp.Pool {
			if worker.Status == WORKING {
				worker.BroadcastUpdateChannel <- global.Signal
			}
		}
	}
}
