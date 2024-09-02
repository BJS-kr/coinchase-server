package task

import (
	"context"
	"log/slog"

	"multiplayer_server/global"
	"multiplayer_server/worker_pool"
	"net"
	"sync"
	"time"

	"github.com/google/uuid"
)

// statusChannel을 인자로 받는 이유
// status update는 성능 최적화와 가독성을 위해 mutex 사용을 최소화해야한다.
// 각 worker가 같은 채널을 바라보고 있다면 game map이 보다 빠르게 처리할 수 있을 것이다.
func LaunchWorkers(workerCount int, statusChannel chan *global.Status) {
	workerPool := worker_pool.GetWorkerPool()
	var initWorker sync.WaitGroup

	// main goroutine이 직접 요청을 받기전 WORKER_COUNT만큼 워커를 활성화
	if len(workerPool.Pool) >= worker_pool.WORKER_COUNT {
		slog.Debug("worker pool is already full")
		return
	}

	for workerId := 0; workerId < workerCount; workerId++ {

		mutualTerminationSignal := make(chan bool)

		// Add를 워커 시작전에 호출하는 이유는 Done이 Add보다 먼저 호출되는 경우를 막기 위해서이다.
		initWorker.Add(2)

		tcpListener := MakeTCPListener()
		port := tcpListener.Addr().(*net.TCPAddr).Port
		worker := workerPool.MakeWorker(port)

		sendMutualTerminationSignal := CollectWorkerForMutualTermination(worker)
		worker.CollectedSendUserRelatedDataToClient = CollectToSendUserRelatedDataToClient(mutualTerminationSignal, sendMutualTerminationSignal, time.Millisecond*100)
		workerPool.Put(uuid.New().String(), worker)

		go ReceiveDataFromClient(tcpListener, statusChannel, &initWorker, mutualTerminationSignal, sendMutualTerminationSignal)
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
		close(workerInitializationSuccessSignal)
	}
}
