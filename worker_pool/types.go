package worker_pool

import (
	"coin_chase/game"
	"coin_chase/protodef"
	"net"
	"sync"
)

type WorkerStatus int

type WorkerPool struct {
	mtx         sync.Mutex
	Pool        map[string]*Worker
	Initialized bool
}

type Worker struct {
	ClientIP                             *net.IP
	ClientPort                           int
	Port                                 int
	OwnerUserID                          string
	Status                               WorkerStatus
	StatusReceiver                       <-chan *protodef.Status
	HealthChecker                        chan game.EmptySignal
	ForceExitSignal                      chan game.EmptySignal
	StopClientSendSignal                 chan game.EmptySignal
	CollectedSendUserRelatedDataToClient func(clientID string, clientIP *net.IP, clientPort int, stopClientSendSignal chan game.EmptySignal)
	BroadcastUpdateChannel               chan game.EmptySignal
}
