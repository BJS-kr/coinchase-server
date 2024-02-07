package main

import (
	"fmt"
	"io"
	"log"
	"multiplayer_server/protodef"
	"net"
	"net/http"
	"sync"
	"time"

	"google.golang.org/protobuf/proto"
)


const WORKER_COUNT int = 10

type Job struct{
	Id int32
	X int32
	Y int32
	Items []int32
	SentAt time.Time
}

type ChannelAndPort struct {
	WorkerID    int
	JobReceiver <-chan Job
	Port        int
}

// graceful shutdown(wait until return이나 terminate signal(runtime.Goexit)등)을 만들지 않은 이유
// main goroutine이 종료된다고 해서 나머지 goroutine이 동시에 처리되는 것은 아니나, 이는 leak을 만들지 않고 결국 종료된다.
// 자세한 내용은 https://stackoverflow.com/questions/72553044/what-happens-to-unfinished-goroutines-when-the-main-parent-goroutine-exits-or-re을 참고
func receiveDataFromClientAndSendJob(conn *net.UDPConn, jobSender chan <- Job, initWorker *sync.WaitGroup) {
	defer conn.Close()
	initWorker.Done()

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
			Id:status.Id,
			X: status.X,
			Y: status.X,
			Items: status.Items,
			SentAt: status.SentAt.AsTime(),
		}
	}
}

func work(workerId int, port int, initWorker *sync.WaitGroup, jobReceiver <-chan Job, workerFunnel chan<- ChannelAndPort, disconnectSignal <-chan bool) {
	workerData := ChannelAndPort{
		WorkerID:    workerId,
		JobReceiver: jobReceiver,
		Port:        port,
	}

	workerFunnel <- workerData
	initWorker.Done()

	for {
		select {
		case job := <-jobReceiver:
			// not ok를 단순히 log로 처리하는 이유는 일정 정도의 데이터 누락을 무시하는
			// UDP기반 데이터 정합성 처리의 특성을 따라 처리 과정에서 무시하기 위함이다.
			println(job)
		case <-disconnectSignal:
			workerFunnel <- workerData
		}
	}
}

func main() {
	workerPool := make([]ChannelAndPort, 0, WORKER_COUNT)
	workerFunnel := make(chan ChannelAndPort)
	disconnectSignal := make(chan bool)
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
		go work(workerId, port, &initWorker, jobChannel, workerFunnel, disconnectSignal)
	}

	initWorker.Wait()

	if len(workerPool) != WORKER_COUNT {
		panic("worker initialization failed")
	} else {
		println("worker initialization succeeded. main goroutine now receiving assign request")
	}

	// Go에서 protobuf를 사용하기 위해 필요한 단계: https://protobuf.dev/getting-started/gotutorial/
	// ex) protoc --go_out=$PWD proto/status.proto
	http.HandleFunc("/get-worker-port", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, "invalid request")

			return 
		}
		var channelAndPort ChannelAndPort
		var err error

		workerPool, channelAndPort, err = Pop(workerPool)

		w.Header().Set("Content-Type", "text/plain")

		if err != nil {
			w.WriteHeader(http.StatusConflict)
			io.WriteString(w, "worker not available")

			return
		}

		w.WriteHeader(http.StatusOK)
		io.WriteString(w, fmt.Sprintf("%d", channelAndPort.Port))
	})

	http.HandleFunc("/disconnect", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, "invalid request")
			
			return
		}

		
	})

	log.Fatal(http.ListenAndServe(":8888", nil))
}
