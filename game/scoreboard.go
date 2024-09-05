package game

func SetUserScore(userId string, score int32) {
	scoreboard[userId] = score
}

func DeleteUserFromScoreboard(userId string) {
	delete(scoreboard, userId)
}

func GetScoreboard() map[string]int32 {
	return scoreboard
}
