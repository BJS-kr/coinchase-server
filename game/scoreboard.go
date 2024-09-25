package game

func (sc *Scoreboard) SetUserScore(userId string, score int32) {
	sc.rwmtx.Lock()
	defer sc.rwmtx.Unlock()

	sc.board[userId] = score
}

func (sc *Scoreboard) IncreaseUserScore(userId string, amount int32) {
	sc.rwmtx.Lock()
	defer sc.rwmtx.Unlock()

	sc.board[userId] += amount
}

func (sc *Scoreboard) DecreaseUserScore(userId string, amount int32) {
	sc.rwmtx.Lock()
	defer sc.rwmtx.Unlock()

	if sc.board[userId] < amount {
		sc.board[userId] = 0
	} else {
		sc.board[userId] -= amount
	}
}

func (sc *Scoreboard) DeleteUserFromScoreboard(userId string) {
	sc.rwmtx.Lock()
	defer sc.rwmtx.Unlock()

	delete(sc.board, userId)
}

func (sc *Scoreboard) GetCopiedBoard() map[string]int32 {
	copiedBoard := make(map[string]int32)

	sc.rwmtx.RLock()
	defer sc.rwmtx.RUnlock()

	for userId, score := range sc.board {
		copiedBoard[userId] = score
	}

	return copiedBoard
}

func GetScoreboard() *Scoreboard {
	return &scoreboard
}
