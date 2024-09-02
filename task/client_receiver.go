package task

import (
	"log"
	"multiplayer_server/global"
	"multiplayer_server/protodef"
	"time"

	"net"
	"sync"

	"log/slog"

	"google.golang.org/protobuf/proto"
)

// Go에서 protobuf를 사용하기 위해 필요한 단계: https://protobuf.dev/getting-started/gotutorial/
// ex) protoc --go_out=$PWD protodef/status.proto

// graceful shutdown(wait until return이나 terminate signal(runtime.Goexit)등)을 만들지 않은 이유
// main goroutine이 종료된다고 해서 나머지 goroutine이 동시에 처리되는 것은 아니나, 이는 leak을 만들지 않고 결국 종료된다.
// 자세한 내용은 https://stackoverflow.com/questions/72553044/what-happens-to-unfinished-goroutines-when-the-main-parent-goroutine-exits-or-re을 참고
const KEEPALIVE_WAIT_LIMIT = time.Second * 300

func ReceiveDataFromClient(tcpListener *net.TCPListener, statusSender chan<- *global.Status, initWorker *sync.WaitGroup, mutualTerminationSignal chan bool, sendMutualTerminationSignal func(chan bool)) {
	defer sendMutualTerminationSignal(mutualTerminationSignal)
	defer tcpListener.Close()

	initWorker.Done()

	slog.Info("Client receiver initialized")
	// IPv4체계에서 최소 패킷의 크기는 576bytes이다(https://networkengineering.stackexchange.com/questions/76459/what-is-the-minimum-mtu-of-ipv4-68-bytes-or-576-bytes#:~:text=576%20bytes%20is%20the%20minimum%20IPv4%20packet%20(datagram)%20size%20that,must%20be%20able%20to%20handle).
	// 이 중 헤더를 뺀 값이 508bytes이며, 이는 UDP라 할지라도 절대 나뉘어질 수 없는 최소크기이다.
	// 그러나 일반적으로 2의 제곱수를 할당하는 것이 관례이므로 576보다 큰 최소 2의 제곱수 1024로 buffer를 만든다.
	// TODO keepalive로 수신하자
	buffer := make([]byte, 1024)

	for {
		select {
		case <-mutualTerminationSignal:
			slog.Info("Termination signal receive in TCP receiver")
			return

		default:
			conn, err := tcpListener.AcceptTCP()

			if err != nil {
				log.Fatal("custom message: TCP accepting failed\n" + err.Error())
			}

			// 한 번 클라이언트와 커넥션이 연결되면 업데이트를 별도의 커넥션으로 받지 않고 keepalive로 계속 받는다.
			// TODO 테스트 필요. 연결이 계속 체결되어있든 말든 일단 유저가 업데이트 패킷을 한번 보내면 거기에 대한 상태 업데이트를 실행해야한다. keepalive상태일때도 이게 계속 되는지를 테스트해봐야함
			err = conn.SetKeepAlive(true)

			if err != nil {

			}

			err = conn.SetKeepAlivePeriod(KEEPALIVE_WAIT_LIMIT)

			if err != nil {

			}

			amount, err := conn.Read(buffer)

			if err != nil {
				log.Fatal("custom message: Read from TCP connection failed\n" + err.Error())
			}

			protoStatus := new(protodef.Status)
			err = proto.Unmarshal(buffer[:amount], protoStatus)

			if err != nil {
				log.Fatal("custom message: TCP unmarshal failed\n" + err.Error())
			}

			userStatus := &global.Status{
				Type: global.STATUS_TYPE_USER,
				Id:   protoStatus.Id,
				CurrentPosition: global.Position{
					X: protoStatus.CurrentPosition.X,
					Y: protoStatus.CurrentPosition.Y,
				},
			}

			statusSender <- userStatus
			// 성능을 위해 buffer를 재사용한다.
			// buffer에 nil을 할당하게 되면 underlying array가 garbage collection되므로 단순히 slice의 길이를 0으로 만든다.
			// 고려사항에 ring buffer가 있었으나, container/ring이 성능적으로 더 나은지 테스트를 해보지 않아 일단 직관적인 구현
			buffer = buffer[:0]
			err = conn.Close()

			if err != nil {
				log.Fatal("custom message: closing TCP connection failed\n" + err.Error())
			}
		}
	}
}
