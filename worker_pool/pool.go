package worker_pool

import (
	"coin_chase/game"
	"coin_chase/worker_pool/worker_status"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"sync"
	"time"

	"github.com/google/uuid"
)

func GetWorkerPool() *WorkerPool {
	return &workerPool
}

func (wp *WorkerPool) GetCopiedPool() Pool {
	wp.rwmtx.RLock()
	defer wp.rwmtx.RUnlock()

	newPool := make(Pool)
	for workerId, worker := range wp.Pool {
		newPool[workerId] = worker
	}

	return newPool
}

const WORKER_INIT_TIMEOUT_SEC = time.Second * 3

// statusChannel을 인자로 받는 이유
// status update는 성능 최적화와 가독성을 위해 mutex 사용을 최소화해야한다.
// 각 worker가 같은 채널을 바라보고 있다면 game map이 보다 빠르게 처리할 수 있을 것이다.
func (wp *WorkerPool) LaunchWorkers(workerCount int, statusChannel chan *game.Status, maximumWorkerCount int) {

	var initWorker sync.WaitGroup

	// main goroutine이 직접 요청을 받기전 WORKER_COUNT만큼 워커를 활성화
	if len(wp.Pool) >= maximumWorkerCount {
		slog.Debug("worker pool is already full")
		return
	}

	for workerId := 0; workerId < workerCount; workerId++ {
		initWorker.Add(1)

		tcpListener := MakeTCPListener()
		port := tcpListener.Addr().(*net.TCPAddr).Port
		worker := wp.MakeWorker(port)

		broadcastUpdateChannel := make(chan game.EmptySignal)
		mutualTerminationContext, mutualCancel := context.WithCancel(context.Background())
		sendMutualTerminationSignal := CollectWorkerForMutualTermination(worker, mutualCancel)

		worker.SendUserRelatedDataToClient = worker.CollectToSendUserRelatedDataToClient(sendMutualTerminationSignal, mutualTerminationContext, broadcastUpdateChannel)
		worker.BroadcastUpdateChannel = broadcastUpdateChannel
		workerId := uuid.New().String()
		wp.Put(workerId, worker)

		go worker.ReceiveDataFromClient(tcpListener, statusChannel, &initWorker, sendMutualTerminationSignal, mutualTerminationContext)
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
		close(workerInitializationSuccessSignal)
		slog.Info("worker initialization succeeded")
	}
}

func (wp *WorkerPool) GetAvailableWorkerCount() int {
	count := 0

	for _, worker := range wp.Pool {
		if worker.GetStatus() == worker_status.IDLE {
			count++
		}
	}

	return count
}

func (wp *WorkerPool) Pull() (*Worker, error) {
	wp.rwmtx.Lock()

	defer wp.rwmtx.Unlock()

	for _, worker := range wp.Pool {
		if worker.GetStatus() == worker_status.IDLE {
			worker.ChangeStatus(worker_status.PULLED_OUT)

			slog.Info("Worker Pulled Out")

			return worker, nil
		}
	}

	return nil, errors.New("worker currently not available")
}

func (wp *WorkerPool) Put(workerId string, worker *Worker) error {
	wp.rwmtx.Lock()

	defer wp.rwmtx.Unlock()

	if worker.GetStatus() != worker_status.WORKING && worker.GetStatus() != worker_status.IDLE {
		worker.ForceExitSignal <- game.Signal
		slog.Debug("INVALID STATUS CHANGE: WORKER STATUS NOT \"WORKING\" OR \"IDLE\"")

		return errors.New("INVALID STATUS CHANGE: WORKER STATUS NOT \"WORKING\" OR \"IDLE\"")
	}

	if worker.GetStatus() == worker_status.WORKING {
		worker.StopClientSendSignal <- game.Signal
	}

	worker.ChangeStatus(worker_status.IDLE)
	worker.OwnerUserID = ""
	worker.ClientIP = nil
	worker.ClientPort = 0

	wp.Pool[workerId] = worker

	slog.Info("Put Worker to pool")

	return nil
}

func (wp *WorkerPool) Delete(workerId string) {
	wp.rwmtx.Lock()

	defer wp.rwmtx.Unlock()

	delete(wp.Pool, workerId)

}

func (wp *WorkerPool) PoolSize() int {
	return len(wp.Pool)
}

func (wp *WorkerPool) GetWorkerByUserId(userId string) (string, *Worker, error) {
	for workerId, worker := range wp.Pool {
		if worker.OwnerUserID == userId {
			return workerId, worker, nil
		}
	}

	return "", nil, errors.New("worker not found")
}

func (wp *WorkerPool) MakeWorker(port int) *Worker {
	return &Worker{
		Port:                 port,
		status:               worker_status.IDLE,
		HealthChecker:        make(chan game.EmptySignal),
		ForceExitSignal:      make(chan game.EmptySignal),
		StopClientSendSignal: make(chan game.EmptySignal),
	}
}

func (wp *WorkerPool) BroadcastSignal(broadcastChannel chan game.EmptySignal) {
	for range broadcastChannel {
		for _, worker := range wp.GetCopiedPool() {
			if worker.GetStatus() == worker_status.WORKING {
				worker.BroadcastUpdateChannel <- game.Signal
			}
		}
	}
}
