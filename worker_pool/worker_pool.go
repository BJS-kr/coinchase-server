package worker_pool

import (
	"errors"
	"sync"
	"time"
)
type Job struct {
	Id     int32
	X      int32
	Y      int32
	Items  []int32
	SentAt time.Time
}
type WorkerPool struct {
	mtx sync.Mutex
	pool map[int]Worker
}
type ChannelAndPort struct {
	JobReceiver <-chan Job
	Port        int
}

type Worker struct {
	ChannelAndPort
	Working bool
}
func (wp *WorkerPool)Pull() (*Worker, error){
	wp.mtx.Lock()

	defer wp.mtx.Unlock()

	for workerId, worker := range wp.pool {
		if !worker.Working {
			worker.Working = true
			wp.pool[workerId] = worker

			return &worker, nil
		}
	}

	return nil, errors.New("worker currently not available")
}

func (wp *WorkerPool)Put(workerId int, worker Worker) {
	wp.mtx.Lock()
	wp.pool[workerId] = worker
	wp.mtx.Unlock()
}

func (wp *WorkerPool)PoolSize() int {
	return len(wp.pool)
}

func (wp *WorkerPool)GetWorkerById(workerId int) (Worker, bool) {
	worker, ok := wp.pool[workerId]	

	return worker, ok
}

func (wp *WorkerPool)Initialize()  {
	wp.pool = make(map[int]Worker)
}