package game

import "coin_chase/game/item_effects"

func GetUserStatuses() *UserStatuses {
	if !userStatuses.Initialized {
		userStatuses.StatusMap = make(map[string]*UserStatus)
		userStatuses.Initialized = true
	}

	return &userStatuses
}

func (uss *UserStatuses) GetUserPosition(userId string) (*Position, bool) {
	userStatus, ok := uss.StatusMap[userId]

	if !ok {
		return nil, ok
	}

	return userStatus.Position, ok
}

func (uss *UserStatuses) SetUserPosition(userId string, X, Y int32) {
	userStatus, ok := uss.StatusMap[userId]

	if !ok {
		uss.StatusMap[userId] = &UserStatus{
			ItemEffect: item_effects.NONE,
			Position: &Position{
				X: X,
				Y: Y,
			},
		}

		return
	}

	uss.StatusMap[userId] = &UserStatus{
		ItemEffect: userStatus.ItemEffect,
		Position: &Position{
			X: X,
			Y: Y,
		},
		ResetTimer: userStatus.ResetTimer,
	}
}
