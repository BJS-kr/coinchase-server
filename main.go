package main

import (
	"fmt"
	"io"
	"log"
	"multiplayer_server/task"
	"multiplayer_server/worker_pool"

	"net/http"
)

const WORKER_COUNT int = 10

func main() {
	// initializeWorkers
	task.LaunchWorkers(WORKER_COUNT)

	if workerPool := worker_pool.GetWorkerPool(); len(workerPool.Pool) != WORKER_COUNT {
		panic("worker pool initialization failed")
	}

	// worker health check
	go task.HealthCheckAndRevive(10)

	// Go에서 protobuf를 사용하기 위해 필요한 단계: https://protobuf.dev/getting-started/gotutorial/
	// ex) protoc --go_out=$PWD proto/status.proto
	http.HandleFunc("GET /get-worker-port/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		workerPool := worker_pool.GetWorkerPool()
		worker, err := workerPool.Pull()
		if err != nil {
			w.WriteHeader(http.StatusConflict)
			io.WriteString(w, "worker currently not available")

			return
		}

		w.WriteHeader(http.StatusOK)
		io.WriteString(w, fmt.Sprintf("%d", worker.Port))
	})

	http.HandleFunc("PATCH /disconnect/{workerId}/", func(w http.ResponseWriter, r *http.Request) {
		workerId := r.PathValue("workerId")

		if workerId == "" {
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, "worker id did not received")
			return
		}
		workerPool := worker_pool.GetWorkerPool()
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
