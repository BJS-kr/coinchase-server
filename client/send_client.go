package client

import (
	"coin_chase/game"
	"coin_chase/protodef"
	"context"
	"fmt"
	"log"
	"log/slog"
	"net"

	"github.com/golang/snappy"
	"google.golang.org/protobuf/proto"
)

func CollectToSendUserRelatedDataToClient(sendMutualTerminationSignal func(), mutualTerminationContext context.Context, broadcastUpdateChannel chan game.EmptySignal) func(clientID string, clientIP *net.IP, clientPort int, stopClientSendSignal chan game.EmptySignal) {
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
					Scoreboard:       gameMap.Scoreboard,
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
