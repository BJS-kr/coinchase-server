package worker_pool

import (
	"errors"

	"sync"
	"time"
)

type Position struct {
	X int32
	Y int32
}

type Item struct {
	Id     string
	Name   string
	Amount int32
}

// 사실상 protodef.Status와 같지만 더 적은 데이터를 넘길 수 있도록 명시
type Status struct {
	Id                string
	CurrentPosition   Position
	Items             []Item
	LastValidPosition Position
	SentAt            time.Time
}
type WorkerPool struct {
	mtx         sync.Mutex
	Pool        map[string]Worker
	Initialized bool
}

type Worker struct {
	JobReceiver     <-chan Status
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

func (wp *WorkerPool) MakeWorker(jobReceiver <-chan Status, port int) Worker {
	return Worker{
		JobReceiver:     jobReceiver,
		Port:            port,
		Working:         false,
		HealthChecker:   make(chan bool),
		ForceExitSignal: make(chan bool),
	}
}
