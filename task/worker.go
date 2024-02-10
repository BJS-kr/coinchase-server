package task

import (
	"multiplayer_server/worker_pool"
	"net"
	"sync"

	"github.com/google/uuid"
)

func Worker(conn *net.UDPConn, initWorker *sync.WaitGroup, jobReceiver <-chan worker_pool.Job, workerPool *worker_pool.WorkerPool, mutualTerminationSignal chan bool) {
	defer SendMutualTerminationSignal(mutualTerminationSignal)

	port := conn.LocalAddr().(*net.UDPAddr).Port
	worker := workerPool.MakeWorker(jobReceiver, port)
	workerPool.Put(uuid.New().String(), worker)

	initWorker.Done()
	println("worker initialized")

	for {
		select {
		case job := <-jobReceiver:
			{

			}
		case <-worker.ForceExitSignal:
			// panic하는 이유는 mutual termination을 실행해야하기 때문이다.
			// 이에 따라 자동으로 UDP Receiver도 종료될 것이다.
			panic("forced exit occurred by signal")

		case <-worker.HealthChecker:
			worker.HealthChecker <- true

		case <-mutualTerminationSignal:
			// 이 시그널이 도착했다는 것은 UDP receiver가 이미 panic했다는 의미이다. 그러므로 단순히 return하면 된다.
			return
		}
	}
}
