package game

import (
	"time"
)

func SendCoinMoveSignalIntervally(coinMoveIntervalMillis int) {
	ticker := time.NewTicker(time.Millisecond * time.Duration(coinMoveIntervalMillis))
	signal := EmptySignal{}

	for range ticker.C {
		CoinMoveSignal <- signal
	}
}
