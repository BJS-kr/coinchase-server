package main

import (
	"fmt"
	"io"
	"log"
	"multiplayer_server/task"
	"multiplayer_server/worker_pool"
	"net"
	"strconv"

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

	http.HandleFunc("GET /get-worker-port/{userId}/{clientIP}/{clientPort}", func(w http.ResponseWriter, r *http.Request) {
		userId := r.PathValue("userId")
		clientIP := net.ParseIP(r.PathValue("clientIP"))
		clientPort, err := strconv.Atoi(r.PathValue("clientPort"))

		w.Header().Set("Content-Type", "text/plain")

		if clientIP == nil || err != nil || userId == "" {
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, "client information invalid")

			return
		}

		workerPool := worker_pool.GetWorkerPool()
		worker, err := workerPool.Pull()

		if err != nil {
			w.WriteHeader(http.StatusConflict)
			io.WriteString(w, "worker currently not available")

			return
		}

		worker.SetClientInformation(userId, &clientIP, clientPort)

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

		workerPool.Put(workerId, worker)

		w.WriteHeader(http.StatusOK)
		io.WriteString(w, "worker successfully returned to pool")
	})

	log.Fatal(http.ListenAndServe(":8888", nil))
}
