package game_map

import (
	"multiplayer_server/protodef"
	"sync"
	"time"
)

type RWMutexGameMap struct {
	Map   *protodef.GameMap
	RWMtx sync.RWMutex
}

var GameMap RWMutexGameMap
var UserPositions map[string]*protodef.Position

// update와 read가 한 곳에서 일어나면 사실상 read가 wlock의 통제를 받게 되므로 Mutex를 사용하는 의미가 없다.
// 그러므로 현재 맵 상태를 전달하는 것과 맵의 상태를 업데이트하는 연산은 별개로 이뤄져야한다.
// 업데이트는 데이터를 전달한 의무가 없으므로 반환 값이 없다.

// 구현 초반에 잘못 생각했던 것은, 자신의 상태가 변하지 않았더라도 계속해서 데이터를 보냈다는 점이다.
// 자신은 상태가 변할 때만 서버에 업데이트 요청을 보내면 되고, 중요한 것은 자신의 주위의 상태를 계속해서 업데이트 받는 것이다.
func (m *RWMutexGameMap) UpdateUserPosition(userStatus *protodef.Status) {
	defer m.RWMtx.Unlock()
	m.RWMtx.Lock()

	if m.isDelayedOver(userStatus, 40) ||
		m.isOutOfRange(userStatus) ||
		m.isOccupied(userStatus) {

		return
	}

	UserPositions[userStatus.Id] = &protodef.Position{
		X: userStatus.CurrentPosition.X,
		Y: userStatus.CurrentPosition.Y,
	}
	m.Map.Rows[userStatus.CurrentPosition.X].Cells[userStatus.CurrentPosition.Y].Occupied = true
	m.Map.Rows[userStatus.CurrentPosition.X].Cells[userStatus.CurrentPosition.Y].Owner = userStatus.Id
}

func (m *RWMutexGameMap) GetSharedMap() *protodef.GameMap {
	defer m.RWMtx.RUnlock()

	// RLock은 여러 goroutine이 획득할 수 있으나, WLock(RWMutex.Lock)은 RLock을 잠그므로 클라이언트가 언제나 업데이트된 상태의 맵을 받을 수 있다.
	m.RWMtx.RLock()

	return m.Map
}

func (m *RWMutexGameMap) isOutOfRange(userStatus *protodef.Status) bool {
	return userStatus.CurrentPosition.X > 99 ||
		userStatus.CurrentPosition.Y > 99 ||
		userStatus.CurrentPosition.X < 0 ||
		userStatus.CurrentPosition.Y < 0
}

func (m *RWMutexGameMap) isDelayedOver(userStatus *protodef.Status, ms int64) bool {
	return time.Now().UnixMilli()-userStatus.SentAt.AsTime().UnixMilli() > ms
}

func (m *RWMutexGameMap) isSamePosition(positionDelta *protodef.Position) bool {
	return positionDelta.X == 0 && positionDelta.Y == 0
}

func (m *RWMutexGameMap) isOccupied(userStatus *protodef.Status) bool {
	return m.Map.Rows[userStatus.CurrentPosition.X].Cells[userStatus.CurrentPosition.Y].Occupied
}
