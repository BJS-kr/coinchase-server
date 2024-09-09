package worker_pool

var workerPool WorkerPool = WorkerPool{
	Pool: make(map[string]*Worker),
}
