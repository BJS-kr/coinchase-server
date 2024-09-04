package global

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

type EmptySignal struct{}
