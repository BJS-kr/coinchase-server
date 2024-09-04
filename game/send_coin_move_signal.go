package game

import (
	"time"
)

func SendCoinMoveSignalIntervally(statusSender chan<- *Status, coinMoveIntervalMillis int) {
	ticker := time.NewTicker(time.Millisecond * time.Duration(coinMoveIntervalMillis))
	coinMoveSignal := &Status{
		Type: STATUS_TYPE_COIN,
	}

	for range ticker.C {
		statusSender <- coinMoveSignal
	}
}
