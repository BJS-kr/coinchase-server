package game

import "time"

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
	Type            string
	Id              string
	CurrentPosition Position
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
	Initialized bool
	Map         Map
	Coins       []*Position
	RandomItems []*Position
	Scoreboard  map[string]int32
}

type UserStatuses struct {
	Initialized bool
	StatusMap   map[string]*UserStatus
}
