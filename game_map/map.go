package game_map

import (
	"log/slog"
	"math/rand/v2"
	"slices"
	"sync"
	"time"
)

const (MAP_SIZE int32 = 20
		COIN_COUNT int = 10
		ITEM_COUNT int = 2
		EFFECT_DURATION int = 10
)

type TileKind int32
type ItemEffect int32
const (
	UNKNOWN = iota
	USER
	COIN
	ITEM_LENGTHEN_VISION
	ITEM_SHORTEN_VISION
	GROUND
)

const (
	UNKNOWN_EFFECT = iota
	NONE = 2
	LENGTHEN = 4
	SHORTEN = 1
)



type Cell struct {
	Occupied bool
	Owner    string
	Kind     TileKind
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
	RandomItems []*Position
	Scoreboard map[string]int32
	RWMtx      sync.RWMutex
}

type UserStatus struct {
	Position *Position
	ItemEffect ItemEffect
	ResetTimer *time.Timer
}
type RWMutexUserStatuses struct {
	mtx           sync.RWMutex
	UserStatuses map[string]*UserStatus
}

type RelatedPosition struct {
	Position *Position
	Cell     *Cell
}

var GameMap RWMutexGameMap
var UserStatuses RWMutexUserStatuses

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
}

func (mup *RWMutexUserStatuses) GetUserPosition(userId string) (*Position, bool) {
	mup.mtx.RLock()
	defer mup.mtx.RUnlock()

	userStatus, ok := mup.UserStatuses[userId]

	if !ok {
		return nil, ok
	}

	return userStatus.Position, ok
}

func (mup *RWMutexUserStatuses) SetUserPosition(userId string, X, Y int32) {
	mup.mtx.Lock()
	defer mup.mtx.Unlock()

	userStatus, ok :=mup.UserStatuses[userId]
	
	if !ok {
		mup.UserStatuses[userId] = &UserStatus{
			ItemEffect: NONE,
			Position: &Position{
				X:X,
				Y:Y,
			},
		}

		return
	}

	mup.UserStatuses[userId] = &UserStatus{
		ItemEffect: userStatus.ItemEffect,
		Position: &Position{
			X:X,
			Y:Y,
		},
		ResetTimer: userStatus.ResetTimer,
	}
}

func (m *RWMutexGameMap) UpdateUserPosition(userStatus *Status) {
	if m.isOutOfRange(&userStatus.CurrentPosition) {
		return
	}

	m.RWMtx.Lock()
	defer m.RWMtx.Unlock()

	if m.isOccupied(&userStatus.CurrentPosition) {
		kind := m.Map.Rows[userStatus.CurrentPosition.Y].Cells[userStatus.CurrentPosition.X].Kind
		if kind == COIN {
			// lock을 얻었으니 MoveCoinsRandomly가 Lock을 얻지 못하고 대기해야하므로, 이곳에서의 정합성은 만족된다.
			coinIdx := slices.IndexFunc(m.Coins, func(coinPosition *Position) bool {
				return coinPosition.X == userStatus.CurrentPosition.X && coinPosition.Y == userStatus.CurrentPosition.Y
			})

			m.Coins = append(m.Coins[:coinIdx], m.Coins[coinIdx+1:]...)

			if len(m.Coins) == 0 {
				m.InitializeCoins()
			}

			m.Scoreboard[userStatus.Id] += 1
		} else if kind == ITEM_LENGTHEN_VISION  || kind == ITEM_SHORTEN_VISION {
			if UserStatuses.UserStatuses[userStatus.Id].ResetTimer != nil {
				UserStatuses.UserStatuses[userStatus.Id].ResetTimer.Stop()
			}
			
			UserStatuses.UserStatuses[userStatus.Id].ResetTimer = time.AfterFunc(time.Second * 6, func() {
				UserStatuses.UserStatuses[userStatus.Id].ItemEffect = NONE
			})
			itemIdx := slices.IndexFunc(m.RandomItems, func(itemPosition *Position) bool {
				return itemPosition.X == userStatus.CurrentPosition.X && itemPosition.Y == userStatus.CurrentPosition.Y
			})

			if itemIdx == -1 {
				slog.Debug("Item exists but not found in slice")
			}

			m.RandomItems = append(m.RandomItems[:itemIdx], m.RandomItems[itemIdx+1:]...)

			if len(m.RandomItems) == 0 {
				m.InitializeItems()
			}
		
			if kind == ITEM_LENGTHEN_VISION {
				// UserStatuses를 변조하고 있으나, 변조하는 스레드들이 각자 RWMutexMap의 Lock을 얻어야하므로 상관없다.
				UserStatuses.UserStatuses[userStatus.Id].ItemEffect = LENGTHEN
			} else if kind == ITEM_SHORTEN_VISION{
				UserStatuses.UserStatuses[userStatus.Id].ItemEffect = SHORTEN
			}
		} else {
			return
		}

	}
	// 이 currentPosition은 서버에 저장된 user의 위치 정보로, userStatus.CurrentPosition과는 다른 값이다.
	currentPosition, exists := UserStatuses.GetUserPosition(userStatus.Id)

	if exists {
		m.Map.Rows[currentPosition.Y].Cells[currentPosition.X] = &Cell{
			Occupied: false,
			Owner:    "",
			Kind:     GROUND,
		}

	}

	UserStatuses.SetUserPosition(userStatus.Id, userStatus.CurrentPosition.X, userStatus.CurrentPosition.Y)

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

func (m *RWMutexGameMap) GetRelatedPositions(userPosition *Position, visibleRange int32) []*RelatedPosition {
	surroundedPositions := make([]Position, 0)

	for x := -visibleRange; x <= visibleRange; x++ {
		for y := -visibleRange; y <= visibleRange; y++ {
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
func (m *RWMutexGameMap) InitializeItems() {
	m.RandomItems = make([]*Position, 0)
	// item은 coin과 다르게 항상 ITEM_COUNT만큼 생성되어야 한다.
	toGenerate := ITEM_COUNT

	for toGenerate > 0 {
		x, y := rand.Int32N(MAP_SIZE), rand.Int32N(MAP_SIZE)
		if m.Map.Rows[y].Cells[x].Occupied { continue }

		itemCell := &Cell{
			Occupied: true,
			Owner: "system",
		}

		if generateRandomDirection() == 1 {
			itemCell.Kind = ITEM_LENGTHEN_VISION
		} else {
			itemCell.Kind = ITEM_SHORTEN_VISION
		}
		
		m.Map.Rows[y].Cells[x] = itemCell
		m.RandomItems = append(m.RandomItems, &Position{
			X:x,
			Y:y,
		})
		toGenerate--
	}

}
func (m *RWMutexGameMap) InitializeCoins() {
	m.Coins = make([]*Position, 0)
	for i := 0; i < COIN_COUNT; i++ { // 겹칠 수 있으니 코인의 갯수도 랜덤. 즉 COIN_COUNT보다 적게 생성 될 수도 있다.
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
	ticker := time.NewTicker(time.Millisecond * 500)

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
