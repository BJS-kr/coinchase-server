package task

import (
	"fmt"
	"log/slog"
	"multiplayer_server/game_map"
	"net"
	"time"

	"google.golang.org/protobuf/proto"
)

func CollectToSendUserRelatedDataToClient(mutualTerminationSignal chan bool, interval time.Duration) func(clientID string, clientIP *net.IP, clientPort int, stopClientSendSignal chan bool) {
	// 먼저 공통의 자원을 수집하기 위해 deferred execution으로 처리

	return func(clientId string, clientIP *net.IP, clientPort int, stopClientSendSignal chan bool) {
		defer SendMutualTerminationSignal(mutualTerminationSignal)
		clientAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", clientIP.String(), clientPort))
		
		if err !=nil {
			slog.Debug(err.Error())
			panic(err)
		}

		client, err := net.DialUDP("udp", nil, clientAddr)

		if err != nil {
			slog.Debug(err.Error())
			panic(err)
		}

		// sleep은 너무 비싼 태스크라 tick로 실행함
		ticker := time.NewTicker(interval)

		for {
			select {
			case <-ticker.C:
				{
					// 일단 POC해보기 위해 전체 맵 데이터보냄
					sharedMap := game_map.GameMap.GetSharedMap()
					data, err := proto.Marshal(sharedMap)
					
					if err != nil {
						slog.Debug(err.Error())
						panic(err)
					}

					client.Write(data)
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