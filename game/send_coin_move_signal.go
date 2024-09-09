package game

import (
	"coin_chase/game/status_types"
	"time"
)

func SendCoinMoveSignalIntervally(statusSender chan<- *Status, coinMoveIntervalMillis int) {
	ticker := time.NewTicker(time.Millisecond * time.Duration(coinMoveIntervalMillis))
	coinMoveSignal := &Status{
		Type: status_types.COIN,
	}

	for range ticker.C {
		statusSender <- coinMoveSignal
	}
}
