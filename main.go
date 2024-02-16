package main

import (
	"fmt"
	"io"
	"log"
	"log/slog"
	"multiplayer_server/game_map"
	"multiplayer_server/task"
	"multiplayer_server/worker_pool"
	"net"
	"os"
	"strconv"

	"net/http"
)

const WORKER_COUNT int = 10


func main() {
	var programLevel = new(slog.LevelVar) // Info by default
	h := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: programLevel})
	slog.SetDefault(slog.New(h))
	programLevel.Set(slog.LevelDebug)
	// initializeWorkers
	task.LaunchWorkers(WORKER_COUNT)

	if workerPool := worker_pool.GetWorkerPool(); len(workerPool.Pool) != WORKER_COUNT {
		panic(fmt.Sprintf("worker pool initialization failed. initialized count: %d, expected count: %d", len(workerPool.Pool), WORKER_COUNT))
	}

	// worker health check
	go task.HealthCheckAndRevive(10)

	game_map.GameMap.Map = &game_map.Map{
		Rows: make([]*game_map.Row, game_map.MAP_SIZE),
	}

	for i := 0; i<int(game_map.MAP_SIZE); i++ {
		game_map.GameMap.Map.Rows[i] = &game_map.Row{
			Cells:  make([]*game_map.Cell, game_map.MAP_SIZE),
		}
		for j := 0; j <int(game_map.MAP_SIZE); j++ {
			game_map.GameMap.Map.Rows[i].Cells[j] = &game_map.Cell{}
		}
	}
	game_map.UserPositions.UserPositions = make(map[string]*game_map.Position)

	http.HandleFunc("GET /get-worker-port/{userId}/{clientIP}/{clientPort}", func(w http.ResponseWriter, r *http.Request) {
		userId := r.PathValue("userId")
		clientIP := net.ParseIP(r.PathValue("clientIP"))
		clientPort, err := strconv.Atoi(r.PathValue("clientPort"))

		slog.Info("client information", "userId", userId, "clientIP", clientIP, "clientPort", clientPort)

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

		worker.StartSendUserRelatedDataToClient()

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
