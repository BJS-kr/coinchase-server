package game_map

import (
	"multiplayer_server/protodef"
	"time"
)

type FieldStatus struct {
	Occupied bool
}

var gameMap [100][100]FieldStatus

func DetermineUserPosition(userStatus *protodef.Status) *protodef.Position {
	// condition 1
	if time.Now().UnixMilli()-userStatus.SentAt.AsTime().UnixMilli() > 40 ||
		// condition 2
		userStatus.CurrentPosition.X > 99 ||
		userStatus.CurrentPosition.Y > 99 ||
		userStatus.CurrentPosition.X < 0 ||
		userStatus.CurrentPosition.Y < 0 ||
		// condition 3
		gameMap[userStatus.CurrentPosition.X][userStatus.CurrentPosition.Y].Occupied {
		return userStatus.LastValidPosition
	}

	return userStatus.CurrentPosition
}
