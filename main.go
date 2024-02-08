package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"multiplayer_server/protodef"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"google.golang.org/protobuf/proto"
)

const WORKER_COUNT int = 10

type Job struct {
	Id     int32
	X      int32
	Y      int32
	Items  []int32
	SentAt time.Time
}

type ChannelAndPort struct {
	JobReceiver <-chan Job
	Port        int
}

type Worker struct {
	ChannelAndPort
	Working bool
}

type WorkerPool struct {
	mtx sync.Mutex
	pool map[int]Worker
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

// graceful shutdown(wait until return이나 terminate signal(runtime.Goexit)등)을 만들지 않은 이유
// main goroutine이 종료된다고 해서 나머지 goroutine이 동시에 처리되는 것은 아니나, 이는 leak을 만들지 않고 결국 종료된다.
// 자세한 내용은 https://stackoverflow.com/questions/72553044/what-happens-to-unfinished-goroutines-when-the-main-parent-goroutine-exits-or-re을 참고
func receiveDataFromClientAndSendJob(conn *net.UDPConn, jobSender chan<- Job, initWorker *sync.WaitGroup) {
	defer conn.Close()
	initWorker.Done()

	println("data receiver initialized")
	for {
		buffer := make([]byte, 1024)
		amount, _, err := conn.ReadFromUDP(buffer)

		if err != nil {
			log.Fatal(err.Error())
		}

		status := protodef.Status{}
		desErr := proto.Unmarshal(buffer[:amount], &status)

		if desErr != nil {
			log.Fatal(err.Error())
		}

		jobSender <- Job{
			Id:     status.Id,
			X:      status.X,
			Y:      status.X,
			Items:  status.Items,
			SentAt: status.SentAt.AsTime(),
		}
	}
}

func work(workerId, port int, initWorker *sync.WaitGroup, jobReceiver <-chan Job, workerPool *WorkerPool) {
	channelAndPort := ChannelAndPort{
		JobReceiver: jobReceiver,
		Port:        port,
	}

	worker := Worker{
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

func main() {
	workerPool := WorkerPool{pool: make(map[int]Worker)}
	var initWorker sync.WaitGroup

	// main goroutine이 직접 요청을 받기전 WORKER_COUNT만큼 워커를 활성화
	for workerId := 0; workerId < WORKER_COUNT; workerId++ {
		jobChannel := make(chan Job)
		addr, err := net.ResolveUDPAddr("udp", ":0")

		if err != nil {
			panic(err)
		}

		conn, err := net.ListenUDP("udp", addr)

		if err != nil {
			panic(err)
		}

		port := conn.LocalAddr().(*net.UDPAddr).Port

		// Add를 워커 시작전에 호출하는 이유는 Done이 Add보다 먼저 호출되는 경우를 막기 위해서이다.
		initWorker.Add(2)
		
		go receiveDataFromClientAndSendJob(conn, jobChannel, &initWorker)
		go work(workerId, port, &initWorker, jobChannel, &workerPool)
	}

	workerInitializationTimeout, cancel := context.WithTimeout(context.Background(), time.Second*5)
	workerInitializationSuccessSignal := make(chan bool)

	go func() {
		defer cancel()

		initWorker.Wait()
		workerInitializationSuccessSignal <- true
	}()

	select {
	case <-workerInitializationTimeout.Done():
		panic("worker initialization did not succeeded in 5 seconds")

	case <-workerInitializationSuccessSignal:
		if workerPool.PoolSize() != WORKER_COUNT {
			panic(fmt.Sprintf("unexpected worker count: %d", workerPool.PoolSize()))
		}

		println("worker initialization succeeded")
	}

	// Go에서 protobuf를 사용하기 위해 필요한 단계: https://protobuf.dev/getting-started/gotutorial/
	// ex) protoc --go_out=$PWD proto/status.proto
	http.HandleFunc("GET /get-worker-port/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")

		worker, err := workerPool.Pull()
		if err != nil {
			w.WriteHeader(http.StatusConflict)
			io.WriteString(w, "worker currently not available")

			return
		}

		w.WriteHeader(http.StatusOK)
		io.WriteString(w, fmt.Sprintf("%d", worker.ChannelAndPort.Port))
	})

	http.HandleFunc("PATCH /disconnect/{workerId}/", func(w http.ResponseWriter, r *http.Request) {
		workerId, err :=strconv.Atoi(r.PathValue("workerId"))
		
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, "worker id did not received")
			return
		}
		worker, ok := workerPool.GetWorkerById(workerId)

		if !ok {
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, "received worker id not found in worker pool")
			return
		}

		if !worker.Working {
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, "worker is not working(already in the pool)")
			return
		}

		worker.Working = false
		workerPool.Put(workerId, worker)

		w.WriteHeader(http.StatusOK)
		io.WriteString(w, "worker successfully returned to pool")
	})

	log.Fatal(http.ListenAndServe(":8888", nil))
}
