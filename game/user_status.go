package game

import (
	"coin_chase/game/item_effects"
	"time"
)

func GetUserStatuses() *UserStatuses {
	return &userStatuses
}

func (uss *UserStatuses) GetUserPosition(userId string) (*Position, bool) {
	uss.rwmtx.RLock()
	defer uss.rwmtx.RUnlock()

	userStatus, ok := uss.StatusMap[userId]

	if !ok {
		return nil, ok
	}

	return userStatus.Position, ok
}

func (uss *UserStatuses) SetUserPosition(userId string, X, Y int32) {
	uss.rwmtx.Lock()
	defer uss.rwmtx.Unlock()

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

func (uss *UserStatuses) GetUserStatus(clientId string) *UserStatus {
	uss.rwmtx.RLock()
	defer uss.rwmtx.RUnlock()

	if userStatus, ok := uss.StatusMap[clientId]; ok {

		return &UserStatus{
			Position:   userStatus.Position,
			ItemEffect: userStatus.ItemEffect,
			ResetTimer: userStatus.ResetTimer,
		}
	}

	return nil
}

func (uss *UserStatuses) SetItemEffect(userId string, itemEffect ItemEffect) {
	uss.rwmtx.Lock()
	defer uss.rwmtx.Unlock()

	uss.StatusMap[userId].ItemEffect = itemEffect
}

func (uss *UserStatuses) SetResetTimer(userId string, timer *time.Timer) {
	uss.rwmtx.Lock()
	defer uss.rwmtx.Unlock()

	uss.StatusMap[userId].ResetTimer = timer
}

func (uss *UserStatuses) GetResetTimer(userId string) *time.Timer {
	userStatus := uss.StatusMap[userId]

	if userStatus == nil {
		return nil
	}

	return uss.StatusMap[userId].ResetTimer
}

func (uss *UserStatuses) RemoveUser(userId string) {
	uss.rwmtx.Lock()
	defer uss.rwmtx.Unlock()

	delete(uss.StatusMap, userId)
}
