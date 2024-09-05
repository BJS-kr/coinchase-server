package http_server

import (
	"coin_chase/bootstrap"
	"coin_chase/game"
	"coin_chase/worker_pool"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"strings"
)

func NewServer(initWorkerCount, maximumWorkerCount int) *http.ServeMux {
	gameMap := game.GetGameMap()
	server := http.NewServeMux()
	bootstrap.Run(initWorkerCount, maximumWorkerCount)

	server.HandleFunc("GET /get-worker-port/{userId}/{clientPort}", func(w http.ResponseWriter, r *http.Request) {
		userId := r.PathValue("userId")
		// client port는 request에서 얻을 수 없다. 여기서 수령하는 포트는 클라이언트의 TCP 리스닝 포트이기 때문이다.
		clientPort, err := strconv.Atoi(r.PathValue("clientPort"))

		slog.Info("client information", "userId", userId, "clientPort", clientPort)

		w.Header().Set("Content-Type", "text/plain")

		clientIP := net.ParseIP(strings.Split(r.RemoteAddr, ":")[0])

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

		game.SetUserScore(userId, 0)
	})

	server.HandleFunc("PATCH /disconnect/{userId}", func(w http.ResponseWriter, r *http.Request) {
		userId := r.PathValue("userId")

		workerPool := worker_pool.GetWorkerPool()
		workerId, worker, err := workerPool.GetWorkerByUserId(userId)

		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			io.WriteString(w, "worker not found")
			return
		}

		workerPool.Put(workerId, worker)
		game.DeleteUserFromScoreboard(userId)

		w.WriteHeader(http.StatusOK)
		io.WriteString(w, "worker successfully returned to pool")
	})

	// 단순히 서버 상태를 조회하는 경로
	server.HandleFunc("GET /server-state", func(w http.ResponseWriter, r *http.Request) {
		workerPool := worker_pool.GetWorkerPool()
		workerCount := workerPool.GetAvailableWorkerCount()
		coinCount := gameMap.CountCoins()
		itemCount := gameMap.CountItems()

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, fmt.Sprintf(`{"workerCount": %d, "coinCount": %d, "itemCount": %d}`, workerCount, coinCount, itemCount))
	})

	return server
}
