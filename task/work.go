package task

import (
	"multiplayer_server/worker_pool"
	"sync"
)


func Work(workerId, port int, initWorker *sync.WaitGroup, jobReceiver <-chan worker_pool.Job, workerPool *worker_pool.WorkerPool) {
	channelAndPort := worker_pool.ChannelAndPort{
		JobReceiver: jobReceiver,
		Port:        port,
	}

	worker := worker_pool.Worker{
		ChannelAndPort: channelAndPort,
		Working: false,
	}

	workerPool.Put(workerId, worker)

	initWorker.Done()
	println("worker initialized")

	for job := range jobReceiver {
		println(job.Id)
	}
}