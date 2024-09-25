// mutex의 대부분을 없앨 수 있을 것 같다.
package game

import (
	"coin_chase/game/item_effects"
	"coin_chase/game/owner_kind"
	"errors"

	"log/slog"
	"math/rand/v2"
	"slices"
	"time"
)

const RESET_DURATION = time.Second * 6

var (
	AttackReceiver = make(chan *Attack)
	StatusReceiver = make(chan *Status)
	CoinMoveSignal = make(chan EmptySignal)

	ErrOutOfRange = errors.New("out of range")
	ErrOccupied   = errors.New("occupied")
)

func MakeMap() Map {
	m := Map{
		Rows: make([]*Row, MAP_SIZE),
	}

	for i := 0; i < int(MAP_SIZE); i++ {
		m.Rows[i] = &Row{
			Cells: make([]*Cell, MAP_SIZE),
		}
		for j := 0; j < int(MAP_SIZE); j++ {
			m.Rows[i].Cells[j] = &Cell{
				Kind: owner_kind.GROUND,
			}
		}
	}

	return m
}

func GetGameMap() *GameMap {
	return &gameMap
}

func (m *GameMap) StartUpdateMap(globalMapUpdateChannel chan<- EmptySignal) {
	for {
		select {
		case status := <-StatusReceiver:
			if err := m.HandleUserStatus(status); err != nil {
				slog.Error("user status not updated", "error", err)
			}
		case <-CoinMoveSignal:
			m.MoveCoinsRandomly()
		case attack := <-AttackReceiver:
			m.HandleAttack(attack)
		}

		globalMapUpdateChannel <- Signal
	}
}

func (m *GameMap) HandleUserStatus(status *Status) error {
	if m.isOutOfRange(&status.CurrentPosition) {
		return ErrOutOfRange
	}

	if m.isOccupied(&status.CurrentPosition) {
		kind := m.Map.Rows[status.CurrentPosition.Y].Cells[status.CurrentPosition.X].Kind

		if kind == owner_kind.COIN {
			coinIdx := slices.IndexFunc(m.coins, func(coinPosition *Position) bool {
				return coinPosition.X == status.CurrentPosition.X && coinPosition.Y == status.CurrentPosition.Y
			})

			m.coins = append(m.coins[:coinIdx], m.coins[coinIdx+1:]...)

			if len(m.coins) == 0 {
				m.InitializeCoins()
			}

			scoreboard.IncreaseUserScore(status.Id)
		} else if kind == owner_kind.ITEM_LENGTHEN_VISION || kind == owner_kind.ITEM_SHORTEN_VISION {
			if resetTimer := userStatuses.GetResetTimer(status.Id); resetTimer != nil {
				resetTimer.Stop()
			}

			userStatuses.SetResetTimer(status.Id, time.AfterFunc(RESET_DURATION, func() {
				userStatuses.SetItemEffect(status.Id, item_effects.NONE)
			}))

			itemIdx := slices.IndexFunc(m.randomItems, func(itemPosition *Position) bool {
				return itemPosition.X == status.CurrentPosition.X && itemPosition.Y == status.CurrentPosition.Y
			})

			if itemIdx == -1 {
				slog.Debug("Item exists but not found in slice")
			}

			m.randomItems = append(m.randomItems[:itemIdx], m.randomItems[itemIdx+1:]...)

			if len(m.randomItems) == 0 {
				m.InitializeItems()
			}

			if kind == owner_kind.ITEM_LENGTHEN_VISION {
				// UserStatuses를 변조하고 있으나, 변조하는 스레드들이 각자 RWMutexMap의 Lock을 얻어야하므로 상관없다.
				userStatuses.SetItemEffect(status.Id, item_effects.LENGTHEN)
			} else if kind == owner_kind.ITEM_SHORTEN_VISION {
				userStatuses.SetItemEffect(status.Id, item_effects.SHORTEN)
			}
		} else {
			// 이동하려는 자리를 다른 유저가 선점한 경우
			return ErrOccupied
		}
	}
	// 이 currentPosition은 서버에 저장된 user의 위치 정보로, userStatus.CurrentPosition과는 다른 값이다.
	currentPosition, exists := userStatuses.GetUserPosition(status.Id)

	if exists {
		m.UpdateMap(currentPosition.X, currentPosition.Y, &Cell{
			Occupied: false,
			Owner:    "",
			Kind:     owner_kind.GROUND,
		})
	}

	userStatuses.SetUserPosition(status.Id, status.CurrentPosition.X, status.CurrentPosition.Y)

	m.UpdateMap(status.CurrentPosition.X, status.CurrentPosition.Y, &Cell{
		Occupied: true,
		Owner:    status.Id,
		Kind:     owner_kind.USER,
	})

	return nil
}

func (m *GameMap) GetRelatedPositions(userPosition *Position, visibleRange int32) []*RelatedPosition {
	surroundedPositions := make([]Position, 0)

	for x := -visibleRange; x <= visibleRange; x++ {
		for y := -visibleRange; y <= visibleRange; y++ {
			if x == 0 && y == 0 {
				continue // 자신의 위치임
			}

			surroundedPositions = append(surroundedPositions, Position{
				X: userPosition.X + x,
				Y: userPosition.Y + y,
			})
		}
	}

	relatedPositions := make([]*RelatedPosition, 0)

	m.rwmtx.RLock()
	defer m.rwmtx.RUnlock()

	for _, surroundedPosition := range surroundedPositions {
		if m.isOutOfRange(&surroundedPosition) {
			continue
		}

		relatedPositions = append(relatedPositions, &RelatedPosition{
			Position: &surroundedPosition,
			Cell:     m.Map.Rows[surroundedPosition.Y].Cells[surroundedPosition.X],
		})
	}

	return relatedPositions
}

func (m *GameMap) isOutOfRange(position *Position) bool {
	return position.X > MAP_SIZE-1 ||
		position.Y > MAP_SIZE-1 ||
		position.X < 0 ||
		position.Y < 0
}

func (m *GameMap) isOccupied(position *Position) bool {
	return m.Map.Rows[position.Y].Cells[position.X].Occupied
}
func (m *GameMap) InitializeItems() {
	m.rwmtx.Lock()
	defer m.rwmtx.Unlock()

	m.randomItems = make([]*Position, 0)
	// item은 coin과 다르게 항상 ITEM_COUNT만큼 생성되어야 한다.
	toGenerate := ITEM_COUNT

	for toGenerate > 0 {
		x, y := rand.Int32N(MAP_SIZE), rand.Int32N(MAP_SIZE)
		if m.Map.Rows[y].Cells[x].Occupied {
			continue
		}

		itemCell := &Cell{
			Occupied: true,
			Owner:    OWNER_SYSTEM,
		}

		if GenerateRandomDirection() == 1 {
			itemCell.Kind = owner_kind.ITEM_LENGTHEN_VISION
		} else {
			itemCell.Kind = owner_kind.ITEM_SHORTEN_VISION
		}

		m.Map.Rows[y].Cells[x] = itemCell
		m.randomItems = append(m.randomItems, &Position{
			X: x,
			Y: y,
		})
		toGenerate--
	}
}
func (m *GameMap) InitializeCoins() {
	m.rwmtx.Lock()
	defer m.rwmtx.Unlock()

	m.coins = make([]*Position, 0)
	for i := 0; i < COIN_COUNT; i++ { // 겹칠 수 있으니 코인의 갯수도 랜덤(Occupied되지 않은 곳에만 생성하니까). 즉 COIN_COUNT보다 적게 생성 될 수도 있다.
		x, y := rand.Int32N(MAP_SIZE), rand.Int32N(MAP_SIZE)
		if !m.Map.Rows[y].Cells[x].Occupied {
			m.Map.Rows[y].Cells[x] = &Cell{
				Occupied: true,
				Owner:    OWNER_SYSTEM,
				Kind:     owner_kind.COIN,
			}
			m.coins = append(m.coins, &Position{
				X: x,
				Y: y,
			})
		}
	}
}

func (m *GameMap) CountCoins() int {
	return len(m.coins)
}

func (m *GameMap) CountItems() int {
	return len(m.randomItems)
}

func (m *GameMap) MoveCoinsRandomly() {
	for i, coinPosition := range m.coins {
		newPos := &Position{
			X: coinPosition.X + GenerateRandomDirection(),
			Y: coinPosition.Y + GenerateRandomDirection(),
		}

		if m.isOutOfRange(newPos) || m.isOccupied(newPos) {
			continue
		}

		m.UpdateMap(coinPosition.X, coinPosition.Y, &Cell{
			Occupied: false,
			Owner:    "",
			Kind:     owner_kind.GROUND,
		})

		m.UpdateMap(newPos.X, newPos.Y, &Cell{
			Occupied: true,
			Owner:    OWNER_SYSTEM,
			Kind:     owner_kind.COIN,
		})

		m.coins[i] = newPos
	}
}

func (m *GameMap) UpdateMap(x, y int32, cell *Cell) {
	m.rwmtx.Lock()
	defer m.rwmtx.Unlock()

	m.Map.Rows[y].Cells[x] = cell
}
