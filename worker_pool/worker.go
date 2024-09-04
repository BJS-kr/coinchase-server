package worker_pool

import (
	"coin_chase/game"
	"coin_chase/worker_pool/worker_status"
	"errors"
	"log/slog"
	"net"
)

func (w *Worker) SetClientInformation(userId string, clientIP *net.IP, clientPort int) error {
	if w.Status != worker_status.PULLED_OUT {
		w.ForceExitSignal <- game.Signal
		slog.Debug("INVALID STATUS CHANGE: WORKER STATUS NOT \"IDLE\"")

		return errors.New("INVALID STATUS CHANGE: WORKER STATUS NOT \"IDLE\"")
	}

	w.Status = worker_status.CLIENT_INFORMATION_RECEIVED
	w.OwnerUserID = userId
	w.ClientIP = clientIP
	w.ClientPort = clientPort

	return nil
}

func (w *Worker) StartSendUserRelatedDataToClient() error {
	if w.Status != worker_status.CLIENT_INFORMATION_RECEIVED {
		w.ForceExitSignal <- game.Signal
		slog.Debug("INVALID STATUS CHANGE: WORKER STATUS NOT \"CLIENT_INFORMATION_RECEIVED\"")

		return errors.New("INVALID STATUS CHANGE: WORKER STATUS NOT \"CLIENT_INFORMATION_RECEIVED\"")
	}

	w.Status = worker_status.WORKING

	go w.CollectedSendUserRelatedDataToClient(w.OwnerUserID, w.ClientIP, w.ClientPort, w.StopClientSendSignal)

	return nil
}
