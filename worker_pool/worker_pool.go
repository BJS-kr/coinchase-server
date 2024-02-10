package worker_pool

import (
	"errors"
	"multiplayer_server/protodef"

	"sync"
)

type WorkerPool struct {
	mtx         sync.Mutex
	Pool        map[string]Worker
	Initialized bool
}

type Worker struct {
	StatusReceiver  <-chan *protodef.Status
	Port            int
	Working         bool
	HealthChecker   chan bool
	ForceExitSignal chan bool
}

var workerPool WorkerPool

func GetWorkerPool() *WorkerPool {
	if !workerPool.Initialized {
		workerPool = WorkerPool{
			Pool:        make(map[string]Worker),
			Initialized: true,
		}
	}

	return &workerPool
}

func (wp *WorkerPool) Pull() (*Worker, error) {
	wp.mtx.Lock()

	defer wp.mtx.Unlock()

	for workerId, worker := range wp.Pool {
		if !worker.Working {
			worker.Working = true
			wp.Pool[workerId] = worker

			return &worker, nil
		}
	}

	return nil, errors.New("worker currently not available")
}

func (wp *WorkerPool) Put(workerId string, worker Worker) {
	wp.mtx.Lock()
	wp.Pool[workerId] = worker
	wp.mtx.Unlock()
}

func (wp *WorkerPool) PoolSize() int {
	return len(wp.Pool)
}

func (wp *WorkerPool) GetWorkerById(workerId string) (Worker, bool) {
	worker, ok := wp.Pool[workerId]

	return worker, ok
}

func (wp *WorkerPool) MakeWorker(statusReceiver <-chan *protodef.Status, port int) Worker {
	return Worker{
		StatusReceiver:  statusReceiver,
		Port:            port,
		Working:         false,
		HealthChecker:   make(chan bool),
		ForceExitSignal: make(chan bool),
	}
}
