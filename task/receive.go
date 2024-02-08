package task

import (
	"log"
	"multiplayer_server/protodef"
	"multiplayer_server/worker_pool"
	"net"
	"sync"

	"google.golang.org/protobuf/proto"
)

// graceful shutdown(wait until return이나 terminate signal(runtime.Goexit)등)을 만들지 않은 이유
// main goroutine이 종료된다고 해서 나머지 goroutine이 동시에 처리되는 것은 아니나, 이는 leak을 만들지 않고 결국 종료된다.
// 자세한 내용은 https://stackoverflow.com/questions/72553044/what-happens-to-unfinished-goroutines-when-the-main-parent-goroutine-exits-or-re을 참고
func ReceiveDataFromClientAndSendJob(conn *net.UDPConn, jobSender chan<- worker_pool.Job, initWorker *sync.WaitGroup) {
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

		jobSender <- worker_pool.Job{
			Id:     status.Id,
			X:      status.X,
			Y:      status.X,
			Items:  status.Items,
			SentAt: status.SentAt.AsTime(),
		}
	}
}
