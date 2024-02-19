package game_map

import (
	"math/rand/v2"
	"slices"
	"sync"
	"time"
)

const MAP_SIZE int32 = 20

type Kind int32

const (
	UNKNOWN = iota
	USER
	COIN
	GROUND
)

type Cell struct {
	Occupied bool
	Owner    string
	Kind     Kind
}
type Row struct {
	Cells []*Cell
}
type Map struct {
	Rows []*Row
}
type RWMutexGameMap struct {
	Map        *Map
	Coins      []*Position
	ScoreBoard map[string]int
	RWMtx      sync.RWMutex
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
	if m.isOutOfRange(&userStatus.CurrentPosition) {
		return
	}

	m.RWMtx.Lock()
	defer m.RWMtx.Unlock()

	if m.isOccupied(&userStatus.CurrentPosition) {
		if m.Map.Rows[userStatus.CurrentPosition.Y].Cells[userStatus.CurrentPosition.X].Kind != COIN {
			return
		}
		// lock을 얻었으니 MoveCoinsRandomly가 Lock을 얻지 못하고 대기해야하므로, 이곳에서의 정합성은 만족된다.
		coinIdx := slices.IndexFunc(m.Coins, func(coinPosition *Position) bool {
			return coinPosition.X == userStatus.CurrentPosition.X && coinPosition.Y == userStatus.CurrentPosition.Y
		})

		m.Coins = append(m.Coins[:coinIdx], m.Coins[coinIdx+1:]...)

		if len(m.Coins) == 0 {
			m.InitializeCoins()
		}

		m.ScoreBoard[userStatus.Id] += 1
	}
	// 이 currentPosition은 서버에 저장된 user의 위치 정보로, userStatus.CurrentPosition과는 다른 값이다.
	currentPosition, exists := UserPositions.GetUserPosition(userStatus.Id)

	if exists {
		m.Map.Rows[currentPosition.Y].Cells[currentPosition.X] = &Cell{
			Occupied: false,
			Owner:    "",
			Kind:     GROUND,
		}

	}

	UserPositions.SetUserPosition(userStatus.Id, userStatus.CurrentPosition.X, userStatus.CurrentPosition.Y)

	m.Map.Rows[userStatus.CurrentPosition.Y].Cells[userStatus.CurrentPosition.X] = &Cell{
		Occupied: true,
		Owner:    userStatus.Id,
		Kind:     USER,
	}

}

func (m *RWMutexGameMap) GetSharedMap() *Map {
	m.RWMtx.RLock()
	defer m.RWMtx.RUnlock()

	// RLock은 여러 goroutine이 획득할 수 있으나, WLock(RWMutex.Lock)은 RLock을 잠그므로 클라이언트가 언제나 업데이트된 상태의 맵을 받을 수 있다.
	return m.Map
}

func (m *RWMutexGameMap) GetRelatedPositions(userPosition *Position) []*RelatedPosition {
	surroundedPositions := make([]Position, 0)
	var x int32
	var y int32
	abs := int32(3)
	for x = -abs; x <= abs; x++ {
		for y = -abs; y <= abs; y++ {
			if x == 0 && y == 0 {
				continue
			} // 자신의 위치임
			surroundedPositions = append(surroundedPositions, Position{
				X: userPosition.X + x,
				Y: userPosition.Y + y,
			})
		}
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

func (m *RWMutexGameMap) isOccupied(position *Position) bool {
	return m.Map.Rows[position.Y].Cells[position.X].Occupied
}

func (m *RWMutexGameMap) InitializeCoins() {
	m.Coins = make([]*Position, 0)
	for i := 0; i < 10; i++ { // 겹칠 수 있으니 코인의 갯수도 랜덤
		x, y := rand.Int32N(MAP_SIZE), rand.Int32N(MAP_SIZE)
		if !m.Map.Rows[y].Cells[x].Occupied {
			m.Map.Rows[y].Cells[x] = &Cell{
				Occupied: true,
				Owner:    "system",
				Kind:     COIN,
			}
			m.Coins = append(m.Coins, &Position{
				X: x,
				Y: y,
			})
		}
	}
}

func (m *RWMutexGameMap) MoveCoinsRandomly() {
	ticker := time.NewTicker(time.Second)

	for _ = range ticker.C {
		func() {
			m.RWMtx.Lock()
			defer m.RWMtx.Unlock()

			for i, coinPosition := range m.Coins {
				newPos := &Position{
					X: coinPosition.X + generateRandomDirection(),
					Y: coinPosition.Y + generateRandomDirection(),
				}

				if m.isOutOfRange(newPos) || m.isOccupied(newPos) {
					continue
				}

				m.Map.Rows[coinPosition.Y].Cells[coinPosition.X] = &Cell{
					Occupied: false,
					Owner:    "",
					Kind:     GROUND,
				}

				m.Map.Rows[newPos.Y].Cells[newPos.X] = &Cell{
					Occupied: true,
					Owner:    "system",
					Kind:     COIN,
				}

				m.Coins[i] = newPos
			}
		}()
	}
}

func generateRandomDirection() int32 {
	if rand.Int32N(2) == 0 {
		return -1
	}

	return 1
}
