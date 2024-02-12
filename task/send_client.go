package task

import (
	"multiplayer_server/game_map"
	"time"
)

func CollectToSendUserRelatedDataToClient(mutualTerminationSignal chan bool, interval time.Duration) func() {
	// 먼저 공통의 자원을 수집하기 위해 deferred execution으로 처리

	return func() {
		defer SendMutualTerminationSignal(mutualTerminationSignal)

		// sleep은 너무 비싼 태스크라 tick로 실행함
		ticker := time.NewTicker(interval)

		for {
			select {
			case <-ticker.C:
				{
					sharedMap := game_map.GameMap.GetSharedMap()
				}
			case <-mutualTerminationSignal:
				return

			}
		}
	}
}
