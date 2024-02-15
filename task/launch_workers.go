package task

import (
	"context"
	"log/slog"

	"multiplayer_server/protodef"
	"multiplayer_server/worker_pool"
	"net"
	"sync"
	"time"

	"github.com/google/uuid"
)

func LaunchWorkers(workerCount int) {
	workerPool := worker_pool.GetWorkerPool()
	var initWorker sync.WaitGroup

	// main goroutine이 직접 요청을 받기전 WORKER_COUNT만큼 워커를 활성화
	for workerId := 0; workerId < workerCount; workerId++ {
	
		statusChannel := make(chan *protodef.Status)
		mutualTerminationSignal := make(chan bool)

		// Add를 워커 시작전에 호출하는 이유는 Done이 Add보다 먼저 호출되는 경우를 막기 위해서이다.
		initWorker.Add(2)

		conn := MakeUDPConn()
		port := conn.LocalAddr().(*net.UDPAddr).Port
		worker := workerPool.MakeWorker(statusChannel, port)
		
		worker.CollectedSendUserRelatedDataToClient = CollectToSendUserRelatedDataToClient(mutualTerminationSignal, time.Millisecond*100)
		workerPool.Put(uuid.New().String(), worker)
		
		go ReceiveDataFromClient(conn, statusChannel, &initWorker, mutualTerminationSignal)
		go ProcessIncoming(worker, &initWorker, statusChannel, workerPool, mutualTerminationSignal)
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
		slog.Info("worker initialization succeeded")
	}
}
