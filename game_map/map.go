package game_map

import (
	"sync"
	"time"
)

const MAP_SIZE int32 = 20

type Cell struct {
	Occupied bool
	Owner    string
}
type Row struct {
	Cells []*Cell
}
type Map struct {
	Rows []*Row
}
type RWMutexGameMap struct {
	Map   *Map
	RWMtx sync.RWMutex
}
type RWMutexUserPositions struct {
	mtx           sync.RWMutex
	UserPositions map[string]*Position
}

type RelatedPosition struct {
	Position *Position
	Cell     *Cell
}

var GameMap RWMutexGameMap
var UserPositions RWMutexUserPositions

// update와 read가 한 곳에서 일어나면 사실상 read가 wlock의 통제를 받게 되므로 Mutex를 사용하는 의미가 없다.
// 그러므로 현재 맵 상태를 전달하는 것과 맵의 상태를 업데이트하는 연산은 별개로 이뤄져야한다.
// 업데이트는 데이터를 전달한 의무가 없으므로 반환 값이 없다.

// 구현 초반에 잘못 생각했던 것은, 자신의 상태가 변하지 않았더라도 계속해서 데이터를 보냈다는 점이다.
// 자신은 상태가 변할 때만 서버에 업데이트 요청을 보내면 되고, 중요한 것은 자신의 주위의 상태를 계속해서 업데이트 받는 것이다.
type Position struct {
	X int32
	Y int32
}

type Item struct {
	Id     string
	Name   string
	Amount int32
}
type Status struct {
	Id              string
	CurrentPosition Position
	Items           []Item
	SentAt          time.Time
}

func (mup *RWMutexUserPositions) GetUserPosition(userId string) (*Position, bool) {
	mup.mtx.RLock()
	defer mup.mtx.RUnlock()

	position, ok := mup.UserPositions[userId]

	return position, ok
}

func (mup *RWMutexUserPositions) SetUserPosition(userId string, X, Y int32) {
	mup.mtx.Lock()
	defer mup.mtx.Unlock()

	mup.UserPositions[userId] = &Position{
		X: X,
		Y: Y,
	}
}

func (m *RWMutexGameMap) UpdateUserPosition(userStatus *Status) {
	if m.isOutOfRange(&userStatus.CurrentPosition) ||
		m.isOccupied(userStatus) {

		return
	}

	m.RWMtx.Lock()
	defer m.RWMtx.Unlock()

	currentPosition, exists := UserPositions.GetUserPosition(userStatus.Id)

	if exists {
		m.Map.Rows[currentPosition.Y].Cells[currentPosition.X].Occupied = false
		m.Map.Rows[currentPosition.Y].Cells[currentPosition.X].Owner = ""
	}

	UserPositions.SetUserPosition(userStatus.Id, userStatus.CurrentPosition.X, userStatus.CurrentPosition.Y)

	m.Map.Rows[userStatus.CurrentPosition.Y].Cells[userStatus.CurrentPosition.X].Occupied = true
	m.Map.Rows[userStatus.CurrentPosition.Y].Cells[userStatus.CurrentPosition.X].Owner = userStatus.Id
}

func (m *RWMutexGameMap) GetSharedMap() *Map {
	m.RWMtx.RLock()
	defer m.RWMtx.RUnlock()

	// RLock은 여러 goroutine이 획득할 수 있으나, WLock(RWMutex.Lock)은 RLock을 잠그므로 클라이언트가 언제나 업데이트된 상태의 맵을 받을 수 있다.
	return m.Map
}

func (m *RWMutexGameMap) GetRelatedPositions(userPosition *Position) []*RelatedPosition {
	surroundedPositions := [8]Position{
		{ // left top
			X: userPosition.X - 1,
			Y: userPosition.Y - 1,
		},
		{ // left
			X: userPosition.X - 1,
			Y: userPosition.Y,
		},
		{ // left bottom
			X: userPosition.X - 1,
			Y: userPosition.Y + 1,
		},
		{ // top
			X: userPosition.X,
			Y: userPosition.Y - 1,
		},
		{ // bottom
			X: userPosition.X,
			Y: userPosition.Y + 1,
		},
		{ // right top
			X: userPosition.X + 1,
			Y: userPosition.Y - 1,
		},
		{ // right
			X: userPosition.X + 1,
			Y: userPosition.Y,
		},
		{ // right bottom
			X: userPosition.X + 1,
			Y: userPosition.Y + 1,
		},
	}

	relatedPositions := make([]*RelatedPosition, 0)
	for _, surroundedPosition := range surroundedPositions {
		if m.isOutOfRange(&surroundedPosition) {
			continue
		}
		relatedPosition := RelatedPosition{
			Position: &surroundedPosition,
			Cell:     m.Map.Rows[surroundedPosition.Y].Cells[surroundedPosition.X],
		}
		relatedPositions = append(relatedPositions, &relatedPosition)
	}

	return relatedPositions
}

func (m *RWMutexGameMap) isOutOfRange(position *Position) bool {
	return position.X > MAP_SIZE-1 ||
		position.Y > MAP_SIZE-1 ||
		position.X < 0 ||
		position.Y < 0
}

func (m *RWMutexGameMap) isDelayedOver(userStatus *Status, ms int64) bool {
	return time.Now().UnixMilli()-userStatus.SentAt.UnixMilli() > ms
}

func (m *RWMutexGameMap) isOccupied(userStatus *Status) bool {
	return m.Map.Rows[userStatus.CurrentPosition.X].Cells[userStatus.CurrentPosition.Y].Occupied
}
