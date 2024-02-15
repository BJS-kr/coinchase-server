package worker_pool

import (
	"errors"
	"log/slog"
	"multiplayer_server/protodef"
	"net"

	"sync"
)

type WorkerStatus int

const (
	IDLE = iota + 1
	PULLED_OUT
	CLIENT_INFORMATION_RECEIVED
	WORKING
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
	HealthChecker                        chan bool
	ForceExitSignal                      chan bool
	StopClientSendSignal                 chan bool
	CollectedSendUserRelatedDataToClient func(clientID string, clientIP *net.IP, clientPort int, stopClientSendSignal chan bool)
}

func (w *Worker) SetClientInformation(userId string, clientIP *net.IP, clientPort int) {
	if w.Status != PULLED_OUT {
		w.ForceExitSignal <- true
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
		w.ForceExitSignal <- true
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

func (wp *WorkerPool) Put(workerId string, worker *Worker) {
	if worker.Status != WORKING && worker.Status != IDLE {
		worker.ForceExitSignal <- true
		slog.Debug("INVALID STATUS CHANGE: WORKER STATUS NOT \"WORKING\" OR \"IDLE\"")

		return
	}

	if worker.Status == WORKING {
		worker.StopClientSendSignal <- true
	}

	worker.Status = IDLE
	worker.OwnerUserID = ""
	worker.ClientIP = nil
	worker.ClientPort = 0

	wp.Pool[workerId] = worker

	slog.Info("Put Worker to pool")
}

func (wp *WorkerPool) PoolSize() int {
	return len(wp.Pool)
}

func (wp *WorkerPool) GetWorkerById(workerId string) (*Worker, bool) {
	worker, ok := wp.Pool[workerId]

	return worker, ok
}

func (wp *WorkerPool) MakeWorker(port int) *Worker {
	return &Worker{
		Port:                 port,
		Status:               IDLE,
		HealthChecker:        make(chan bool),
		ForceExitSignal:      make(chan bool),
		StopClientSendSignal: make(chan bool),
		//UserID와 ClientIP와 ClientPort는 추후 워커가 유저에게 할당 될 때 설정된다.
	}
}
