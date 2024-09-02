package task

import (
	"multiplayer_server/global"
	"time"
)

func SendCoinMoveSignalIntervally(statusSender chan<- *global.Status, coinMoveIntervalMillis int) {
	ticker := time.NewTicker(time.Millisecond * time.Duration(coinMoveIntervalMillis))
	coinMoveSignal := &global.Status{
		Type: global.STATUS_TYPE_COIN,
	}

	for range ticker.C {
		statusSender <- coinMoveSignal
	}
}
