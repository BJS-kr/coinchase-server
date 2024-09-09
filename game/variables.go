package game

var gameMap GameMap = GameMap{
	Map: MakeMap(),
}
var userStatuses UserStatuses = UserStatuses{
	StatusMap: make(map[string]*UserStatus),
}
var scoreboard Scoreboard = Scoreboard{
	board: make(map[string]int32),
}
var Signal EmptySignal
