package game_map

import (
	"multiplayer_server/protodef"
	"time"
)

type FieldStatus struct {
	Occupied bool
}

var gameMap [100][100]FieldStatus

func DetermineUserPosition(userStatus protodef.Status) *protodef.Position {
	if time.Now().UnixMilli()-userStatus.SentAt.AsTime().UnixMilli() > 40 {
		return userStatus.LastValidPosition
	}

	if userStatus.CurrentPosition.X > 99 || userStatus.CurrentPosition.Y > 99 || userStatus.CurrentPosition.X < 0 || userStatus.CurrentPosition.Y < 0 {
		return userStatus.LastValidPosition
	}

	if gameMap[userStatus.CurrentPosition.X][userStatus.CurrentPosition.Y].Occupied {
		return userStatus.LastValidPosition
	}

	return userStatus.CurrentPosition
}
