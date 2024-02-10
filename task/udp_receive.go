package task

import (
	"log"
	"multiplayer_server/protodef"
	"multiplayer_server/worker_pool"

	"net"
	"sync"

	"google.golang.org/protobuf/proto"
)

// Go에서 protobuf를 사용하기 위해 필요한 단계: https://protobuf.dev/getting-started/gotutorial/
// ex) protoc --go_out=$PWD protodef/status.proto

// graceful shutdown(wait until return이나 terminate signal(runtime.Goexit)등)을 만들지 않은 이유
// main goroutine이 종료된다고 해서 나머지 goroutine이 동시에 처리되는 것은 아니나, 이는 leak을 만들지 않고 결국 종료된다.
// 자세한 내용은 https://stackoverflow.com/questions/72553044/what-happens-to-unfinished-goroutines-when-the-main-parent-goroutine-exits-or-re을 참고
func ReceiveDataFromClientAndSendJob(conn *net.UDPConn, statusSender chan<- worker_pool.Status, initWorker *sync.WaitGroup, mutualTerminationSignal chan bool) {
	defer SendMutualTerminationSignal(mutualTerminationSignal)
	defer conn.Close()

	initWorker.Done()

	println("data receiver initialized")

	for {
		select {
		case <-mutualTerminationSignal:
			return

		default:
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

			itemLen := len(status.Items)
			items := make([]worker_pool.Item, 0, itemLen)

			for i := range itemLen {
				items = append(items, worker_pool.Item{
					Id:     status.Items[i].Id,
					Name:   status.Items[i].Name,
					Amount: status.Items[i].Amount,
				})
			}

			statusSender <- worker_pool.Status{
				Id: status.Id,
				CurrentPosition: worker_pool.Position{
					X: status.CurrentPosition.X,
					Y: status.CurrentPosition.Y,
				},
				LastValidPosition: worker_pool.Position{
					X: status.LastValidPosition.X,
					Y: status.LastValidPosition.Y,
				},
				Items:  items,
				SentAt: status.SentAt.AsTime(),
			}
		}

	}
}
