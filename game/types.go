package game

import (
	"sync"
	"time"
)

type TileKind int32
type ItemEffect int32

type UserStatus struct {
	Position   *Position
	ItemEffect ItemEffect
	ResetTimer *time.Timer
}

type RelatedPosition struct {
	Position *Position
	Cell     *Cell
}

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

type Attack struct {
	UserId         string
	UserPosition   Position
	AttackPosition Position
}

// boolean보다 효율적. 데이터 크기가 0임
type EmptySignal struct{}

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

type GameMap struct {
	Map         Map
	coins       []*Position
	randomItems []*Position
	rwmtx       sync.RWMutex
}

type Scoreboard struct {
	rwmtx sync.RWMutex
	board map[string]int32
}

type UserStatuses struct {
	StatusMap map[string]*UserStatus
	rwmtx     sync.RWMutex
}
