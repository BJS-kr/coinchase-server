package global

type UserStatuses struct {
	UserStatuses map[string]*UserStatus
}

func (uss *UserStatuses) GetUserPosition(userId string) (*Position, bool) {
	userStatus, ok := uss.UserStatuses[userId]

	if !ok {
		return nil, ok
	}

	return userStatus.Position, ok
}

func (uss *UserStatuses) SetUserPosition(userId string, X, Y int32) {
	userStatus, ok := uss.UserStatuses[userId]

	if !ok {
		uss.UserStatuses[userId] = &UserStatus{
			ItemEffect: NONE,
			Position: &Position{
				X: X,
				Y: Y,
			},
		}

		return
	}

	uss.UserStatuses[userId] = &UserStatus{
		ItemEffect: userStatus.ItemEffect,
		Position: &Position{
			X: X,
			Y: Y,
		},
		ResetTimer: userStatus.ResetTimer,
	}
}
