package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"multiplayer_server/task"
	"multiplayer_server/worker_pool"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"
)

const WORKER_COUNT int = 10

func main() {
	workerPool := worker_pool.WorkerPool{}
	workerPool.Initialize()
	var initWorker sync.WaitGroup

	// main goroutine이 직접 요청을 받기전 WORKER_COUNT만큼 워커를 활성화
	for workerId := 0; workerId < WORKER_COUNT; workerId++ {
		jobChannel := make(chan worker_pool.Job)
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
		
		go task.ReceiveDataFromClientAndSendJob(conn, jobChannel, &initWorker)
		go task.Work(workerId, port, &initWorker, jobChannel, &workerPool)
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
