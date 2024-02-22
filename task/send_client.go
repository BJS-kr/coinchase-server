package task

import (
	"fmt"
	"log"
	"log/slog"
	"multiplayer_server/game_map"
	"multiplayer_server/protodef"
	"net"
	"time"

	"github.com/golang/snappy"
	"google.golang.org/protobuf/proto"
)

func CollectToSendUserRelatedDataToClient(mutualTerminationSignal chan bool, sendMutualTerminationSignal func(chan bool), interval time.Duration) func(clientID string, clientIP *net.IP, clientPort int, stopClientSendSignal chan bool) {
	// 먼저 공통의 자원을 수집하기 위해 deferred execution으로 처리
	return func(clientId string, clientIP *net.IP, clientPort int, stopClientSendSignal chan bool) {
		defer sendMutualTerminationSignal(mutualTerminationSignal)
		clientAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", clientIP.String(), clientPort))

		if err != nil {
			log.Fatal(err.Error())
		}

		client, err := net.DialUDP("udp", nil, clientAddr)

		if err != nil {
			slog.Debug(err.Error())
			panic(err)
		}

		// sleep은 너무 비싼 태스크라 tick로 실행함
		ticker := time.NewTicker(interval)
		faultTolerance := 100

		for {
			select {
			case <-ticker.C:
				userStatus, ok := game_map.UserStatuses.UserStatuses[clientId]

				if !ok {
					continue
				}

				relatedPositions := game_map.GameMap.GetRelatedPositions(userStatus.Position, int32(userStatus.ItemEffect))
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
					Scoreboard:       game_map.GameMap.Scoreboard,
				}

				marshaledProtoUserRelatedPositions, err := proto.Marshal(protoUserRelatedPositions)

				if err != nil {
					log.Fatal(err.Error())
				}

				// packet size를 576바이트 이하로 유지하기 위해 snappy를 씁니다.
				compressedUserRelatedPositions := snappy.Encode(nil, marshaledProtoUserRelatedPositions)
				_, err = client.Write(compressedUserRelatedPositions)

				if err != nil {
					slog.Debug(err.Error(), "fault tolerance remain:", faultTolerance)
					faultTolerance--

					// panic은 연관된 모든 자원을 정리하도록 설계되어 있음
					if faultTolerance <=  0 {
						panic(err)
					}
				}

			case <-mutualTerminationSignal:
				return
			// stopClientSendSignal은 client send가 worker가 생성되고 난 뒤, 클라이언트에서 정보를 받으면 내부적으로 실행되기 때문에, 모든 관계를 죽이는(mutual termination)이 아닌
			// client send만 죽일 필요가 있기 때문에 또 다른 시그널이 필요해진다.
			// 이 시그널은 worker가 worker pool에 put될 때 수신된다.
			case <-stopClientSendSignal:
				return
			}

		}
	}
}
