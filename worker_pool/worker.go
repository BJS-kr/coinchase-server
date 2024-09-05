package worker_pool

import (
	"bytes"
	"coin_chase/game"
	"coin_chase/protodef"
	"coin_chase/worker_pool/worker_status"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net"
	"sync"
	"time"

	"github.com/golang/snappy"
	"google.golang.org/protobuf/proto"
)

// Go에서 protobuf를 사용하기 위해 필요한 단계: https://protobuf.dev/getting-started/gotutorial/
// ex) protoc --go_out=$PWD protodef/status.proto

// graceful shutdown(wait until return이나 terminate signal(runtime.Goexit)등)을 만들지 않은 이유
// main goroutine이 종료된다고 해서 나머지 goroutine이 동시에 처리되는 것은 아니나, 이는 leak을 만들지 않고 결국 종료된다.
// 자세한 내용은 https://stackoverflow.com/questions/72553044/what-happens-to-unfinished-goroutines-when-the-main-parent-goroutine-exits-or-re을 참고
const READ_DEADLINE = time.Second * 300
const BUFFER_SIZE = 4096
const BUFFER_DELIMITER byte = '$'

func (w *Worker) SetClientInformation(userId string, clientIP *net.IP, clientPort int) error {
	if w.GetStatus() != worker_status.PULLED_OUT {
		w.ForceExitSignal <- game.Signal
		slog.Debug("INVALID STATUS CHANGE: WORKER STATUS NOT \"IDLE\"")

		return errors.New("INVALID STATUS CHANGE: WORKER STATUS NOT \"IDLE\"")
	}

	w.ChangeStatus(worker_status.CLIENT_INFORMATION_RECEIVED)
	w.OwnerUserID = userId
	w.ClientIP = clientIP
	w.ClientPort = clientPort

	return nil
}

func (w *Worker) StartSendUserRelatedDataToClient() error {
	if w.GetStatus() != worker_status.CLIENT_INFORMATION_RECEIVED {
		w.ForceExitSignal <- game.Signal
		slog.Debug("INVALID STATUS CHANGE: WORKER STATUS NOT \"CLIENT_INFORMATION_RECEIVED\"")

		return errors.New("INVALID STATUS CHANGE: WORKER STATUS NOT \"CLIENT_INFORMATION_RECEIVED\"")
	}

	w.ChangeStatus(worker_status.WORKING)

	go w.SendUserRelatedDataToClient(w.OwnerUserID, w.ClientIP, w.ClientPort, w.StopClientSendSignal)

	return nil
}

func (w *Worker) ReceiveDataFromClient(tcpListener *net.TCPListener, statusSender chan<- *game.Status, initWorker *sync.WaitGroup, sendMutualTerminationSignal func(), mutualTerminationContext context.Context) {
	defer sendMutualTerminationSignal()
	defer tcpListener.Close()

	initWorker.Done()

	slog.Info("Client receiver initialized")
	// IPv4체계에서 최소 패킷의 크기는 576bytes이다(https://networkengineering.stackexchange.com/questions/76459/what-is-the-minimum-mtu-of-ipv4-68-bytes-or-576-bytes#:~:text=576%20bytes%20is%20the%20minimum%20IPv4%20packet%20(datagram)%20size%20that,must%20be%20able%20to%20handle).
	// 이 중 헤더를 뺀 값이 508bytes이며, 이는 UDP라 할지라도 절대 나뉘어질 수 없는 최소크기이다.
	// 그러나 일반적으로 2의 제곱수를 할당하는 것이 관례이므로 576보다 큰 최소 2의 제곱수 1024로 buffer를 만든다.
	buffer := make([]byte, BUFFER_SIZE)
	queueBuffer := bytes.NewBuffer(nil)
	conn, err := tcpListener.AcceptTCP()

	if err != nil {
		log.Fatal("TCP accepting failed\n" + err.Error())
	}

	// READ_DEADLINE만큼 idle상태이면 클라이언트를 유지할 이유가 없다고 판단하고 종료
	// read deadline에 도달시, 아래의 conn.Read에서 error발생
	// Read할 때 단순히 log.Fatal해버리고 있는데, 어차피 이 함수는 관리되고있기 때문에 관련된 goroutine들이 모두 종료되고 새로운 worker가 삽입된다.
	// 아울러 하단의 for select구문의 default에서 유저가 보낸 데이터가 수신되면 read deadline을 5분씩 연장하고 있다.
	if err := conn.SetReadDeadline(time.Now().Add(READ_DEADLINE)); err != nil {
		log.Fatal("failed to set read deadline to TCP connection")
	}

	if err := conn.SetKeepAlive(true); err != nil {
		log.Fatal("failed to set keepalive to TCP connection")
	}

	defer conn.Close()

	for {
		select {
		case <-mutualTerminationContext.Done():
			slog.Info("Termination signal receive in TCP receiver")
			return
		case <-w.ForceExitSignal:
			slog.Info("force exit signal received in ReceiveDataFromClient")
			return

		default:
			// 성능을 위해 buffer를 재사용한다.
			// buffer에 nil을 할당하게 되면 underlying array가 garbage collection되므로 단순히 slice의 길이를 0으로 만든다.
			// 고려사항에 ring buffer가 있었으나, container/ring이 성능적으로 더 나은지 테스트를 해보지 않아 일단 직관적인 구현
			size, err := conn.Read(buffer)

			if err != nil {
				log.Fatal("Read from TCP connection failed\n" + err.Error())
			}

			if size >= BUFFER_SIZE {
				log.Fatal("received TCP packet size exceeded the buffer size")
			}

			if size > 0 {
				queueBuffer.Write(buffer[:size])

				for {
					data, err := queueBuffer.ReadBytes(BUFFER_DELIMITER)

					if err != nil {
						if errors.Is(err, io.EOF) {
							queueBuffer.Write(data)
							break
						} else {
							log.Fatal("ReadBytes returned error other than EOF(unexpected)", err.Error())
						}
					}

					protoStatus := new(protodef.Status)

					if err := proto.Unmarshal(data[:len(data)-1], protoStatus); err != nil {
						log.Fatal("TCP unmarshal failed\n" + err.Error())
					}

					userStatus := &game.Status{
						Type: game.STATUS_TYPE_USER,
						Id:   protoStatus.Id,
						CurrentPosition: game.Position{
							X: protoStatus.CurrentPosition.X,
							Y: protoStatus.CurrentPosition.Y,
						},
					}

					statusSender <- userStatus
				}

				// 0 이상의 패킷 수신마다 갱신
				conn.SetReadDeadline(time.Now().Add(READ_DEADLINE))
			}
		}
	}
}

func (w *Worker) CollectToSendUserRelatedDataToClient(sendMutualTerminationSignal func(), mutualTerminationContext context.Context, broadcastUpdateChannel chan game.EmptySignal) func(clientID string, clientIP *net.IP, clientPort int, stopClientSendSignal chan game.EmptySignal) {
	// 먼저 공통의 자원을 수집하기 위해 deferred execution으로 처리
	return func(clientId string, clientIP *net.IP, clientPort int, stopClientSendSignal chan game.EmptySignal) {
		defer sendMutualTerminationSignal()
		clientAddr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", clientIP.String(), clientPort))

		if err != nil {
			log.Fatal(err.Error())
		}

		client, err := net.DialTCP("tcp", nil, clientAddr)

		if err != nil {
			slog.Debug(err.Error())
			panic(err)
		}

		faultTolerance := 100

		for {
			select {
			case <-mutualTerminationContext.Done():
				slog.Info("Termination signal received in TCP client sender")
				return
				// stopClientSendSignal은 client send와 worker가 생성되고 난 뒤, 클라이언트에서 정보를 받으면 내부적으로 실행되기 때문에, 모든 관계를 죽이는(mutual termination)이 아닌
				// client send만 죽일 필요가 있기 때문에 또 다른 시그널이 필요해진다.
				// 이 시그널은 worker가 worker pool에 되돌아 갈 때 수신된다.
			case <-stopClientSendSignal:
				slog.Info("stop client send signal received")
				return
			case <-w.ForceExitSignal:
				slog.Info("force exit signal received in SendUserRelatedDataToClient")
				return
			case <-broadcastUpdateChannel:
				gameMap, userStatuses := game.GetGameMap(), game.GetUserStatuses()
				userStatus, ok := userStatuses.StatusMap[clientId]

				if !ok {
					continue
				}

				relatedPositions := gameMap.GetRelatedPositions(userStatus.Position, int32(userStatus.ItemEffect))
				protoUserPosition := &protodef.Position{
					X: userStatus.Position.X,
					Y: userStatus.Position.Y,
				}
				protoRelatedPositions := make([]*protodef.RelatedPosition, 0)

				for _, relatedPosition := range relatedPositions {
					protoCell := &protodef.Cell{
						Occupied: relatedPosition.Cell.Occupied,
						Owner:    relatedPosition.Cell.Owner,
						Kind:     int32(relatedPosition.Cell.Kind),
					}
					protoPosition := &protodef.Position{
						X: relatedPosition.Position.X,
						Y: relatedPosition.Position.Y,
					}
					protoRelatedPositions = append(protoRelatedPositions, &protodef.RelatedPosition{
						Cell:     protoCell,
						Position: protoPosition,
					})
				}

				protoUserRelatedPositions := &protodef.RelatedPositions{
					UserPosition:     protoUserPosition,
					RelatedPositions: protoRelatedPositions,
					Scoreboard:       game.GetScoreboard(),
				}

				marshaledProtoUserRelatedPositions, err := proto.Marshal(protoUserRelatedPositions)

				if err != nil {
					log.Fatal(err.Error())
				}

				// packet size 최소화를 위해 snappy를 씁니다.
				compressedUserRelatedPositions := snappy.Encode(nil, marshaledProtoUserRelatedPositions)
				_, err = client.Write(compressedUserRelatedPositions)

				if err != nil {
					slog.Debug(err.Error(), "fault tolerance remain:", faultTolerance)
					faultTolerance--

					// panic은 연관된 모든 자원을 정리하도록 설계되어 있음
					if faultTolerance <= 0 {
						panic(err)
					}
				}

			}
		}
	}
}

func (w *Worker) ChangeStatus(status int) {
	w.rwmtx.Lock()
	defer w.rwmtx.Unlock()

	w.status = WorkerStatus(status)
}

func (w *Worker) GetStatus() WorkerStatus {
	w.rwmtx.RLock()
	defer w.rwmtx.RUnlock()

	return WorkerStatus(w.status)
}
