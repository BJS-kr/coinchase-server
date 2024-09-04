package http_server

import (
	"coin_chase/game"
	"coin_chase/game/owner_kind"
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
	statusChannel := make(chan *game.Status)
	worker_pool.LaunchWorkers(initWorkerCount, statusChannel, maximumWorkerCount)
	workerPool := worker_pool.GetWorkerPool()

	if workerPool.GetAvailableWorkerCount() != initWorkerCount {
		panic(fmt.Sprintf("worker pool initialization failed. initialized count: %d, expected count: %d", len(workerPool.Pool), initWorkerCount))
	}

	gameMapUpdateChannel := make(chan game.EmptySignal)

	go worker_pool.HealthCheckAndRevive(10, statusChannel, maximumWorkerCount)

	gameMap, userStatuses := game.GetGameMap(), game.GetUserStatuses()

	go gameMap.StartUpdateObjectPosition(statusChannel, gameMapUpdateChannel)
	go workerPool.BroadcastgameMapUpdate(gameMapUpdateChannel)

	gameMap.Scoreboard = make(map[string]int32)
	gameMap.Map = &game.Map{
		Rows: make([]*game.Row, game.MAP_SIZE),
	}

	for i := 0; i < int(game.MAP_SIZE); i++ {
		gameMap.Map.Rows[i] = &game.Row{
			Cells: make([]*game.Cell, game.MAP_SIZE),
		}
		for j := 0; j < int(game.MAP_SIZE); j++ {
			gameMap.Map.Rows[i].Cells[j] = &game.Cell{
				Kind: owner_kind.GROUND,
			}
		}
	}

	gameMap.InitializeCoins()
	gameMap.InitializeItems()

	go game.SendCoinMoveSignalIntervally(statusChannel, 500)

	userStatuses.StatusMap = make(map[string]*game.UserStatus)

	server := http.NewServeMux()
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
		gameMap.Scoreboard[userId] = 0 // 굳이 zero value를 할당하는 이유는 0점이라도 표시가 되어야하기 때문
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
		delete(gameMap.Scoreboard, userId)

		w.WriteHeader(http.StatusOK)
		io.WriteString(w, "worker successfully returned to pool")
	})

	// 서버 상태를 조회하기 위한 간단한 핸들러
	server.HandleFunc("GET /server-state", func(w http.ResponseWriter, r *http.Request) {
		workerPool := worker_pool.GetWorkerPool()
		workerCount := workerPool.GetAvailableWorkerCount()
		coinCount := len(gameMap.Coins)
		itemCount := len(gameMap.RandomItems)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		io.WriteString(w, fmt.Sprintf(`{"workerCount": %d, "coinCount": %d, "itemCount": %d}`, workerCount, coinCount, itemCount))
	})

	return server
}
