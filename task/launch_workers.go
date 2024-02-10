package task

import (
	"context"
	"multiplayer_server/worker_pool"
	"sync"
	"time"
)

func LaunchWorkers(workerCount int) {
	workerPool := worker_pool.GetWorkerPool()
	var initWorker sync.WaitGroup

	// main goroutine이 직접 요청을 받기전 WORKER_COUNT만큼 워커를 활성화
	for workerId := 0; workerId < workerCount; workerId++ {
		jobChannel := make(chan worker_pool.Job)
		mutualTerminationSignal := make(chan bool)

		// Add를 워커 시작전에 호출하는 이유는 Done이 Add보다 먼저 호출되는 경우를 막기 위해서이다.
		initWorker.Add(2)
		conn := MakeUDPConn()
		go ReceiveDataFromClientAndSendJob(conn, jobChannel, &initWorker, mutualTerminationSignal)
		go Worker(conn, &initWorker, jobChannel, workerPool, mutualTerminationSignal)
	}

	workerInitializationTimeout, cancel := context.WithTimeout(context.Background(), time.Second*3)
	workerInitializationSuccessSignal := make(chan bool)

	go func() {
		defer cancel()

		initWorker.Wait()
		workerInitializationSuccessSignal <- true
	}()

	select {
	case <-workerInitializationTimeout.Done():
		panic("worker initialization did not succeeded in 3 seconds")

	case <-workerInitializationSuccessSignal:
		println("worker initialization succeeded")
	}
}
