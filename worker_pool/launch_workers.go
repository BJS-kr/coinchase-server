package worker_pool

import (
	"context"
	"fmt"
	"log/slog"

	"coin_chase/client"
	"coin_chase/game"

	"net"
	"sync"
	"time"

	"github.com/google/uuid"
)

const WORKER_INIT_TIMEOUT_SEC = time.Second * 3

// statusChannel을 인자로 받는 이유
// status update는 성능 최적화와 가독성을 위해 mutex 사용을 최소화해야한다.
// 각 worker가 같은 채널을 바라보고 있다면 game map이 보다 빠르게 처리할 수 있을 것이다.
func LaunchWorkers(workerCount int, statusChannel chan *game.Status, maximumWorkerCount int) {
	workerPool := GetWorkerPool()
	var initWorker sync.WaitGroup

	// main goroutine이 직접 요청을 받기전 WORKER_COUNT만큼 워커를 활성화
	if len(workerPool.Pool) >= maximumWorkerCount {
		slog.Debug("worker pool is already full")
		return
	}

	for workerId := 0; workerId < workerCount; workerId++ {
		initWorker.Add(1)

		tcpListener := MakeTCPListener()
		port := tcpListener.Addr().(*net.TCPAddr).Port
		worker := workerPool.MakeWorker(port)

		broadcastUpdateChannel := make(chan game.EmptySignal)
		mutualTerminationContext, mutualCancel := context.WithCancel(context.Background())
		sendMutualTerminationSignal := CollectWorkerForMutualTermination(worker, mutualCancel)

		worker.CollectedSendUserRelatedDataToClient = client.CollectToSendUserRelatedDataToClient(sendMutualTerminationSignal, mutualTerminationContext, broadcastUpdateChannel)
		worker.BroadcastUpdateChannel = broadcastUpdateChannel

		workerPool.Put(uuid.New().String(), worker)

		go client.ReceiveDataFromClient(tcpListener, statusChannel, &initWorker, sendMutualTerminationSignal, mutualTerminationContext)
	}

	workerInitializationTimeout, workerInitializationTimeoutCancel := context.WithTimeout(context.Background(), WORKER_INIT_TIMEOUT_SEC)
	workerInitializationSuccessSignal := make(chan game.EmptySignal)

	go func() {
		defer workerInitializationTimeoutCancel()

		initWorker.Wait()
		workerInitializationSuccessSignal <- game.Signal
	}()

	select {
	case <-workerInitializationTimeout.Done():
		panic(fmt.Sprintf("worker initialization did not succeeded in %d seconds", WORKER_INIT_TIMEOUT_SEC/1_000_000_000))

	case <-workerInitializationSuccessSignal:
		slog.Info("worker initialization succeeded")
		close(workerInitializationSuccessSignal)
	}
}
