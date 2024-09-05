package worker_pool

import (
	"coin_chase/game"
	"coin_chase/protodef"
	"net"
	"sync"
)

type WorkerStatus int
type Pool map[string]*Worker
type WorkerPool struct {
	rwmtx       sync.RWMutex
	Pool        Pool
	Initialized bool
}

type Worker struct {
	rwmtx                       sync.RWMutex
	ClientIP                    *net.IP
	ClientPort                  int
	Port                        int
	OwnerUserID                 string
	status                      WorkerStatus
	StatusReceiver              <-chan *protodef.Status
	HealthChecker               chan game.EmptySignal
	ForceExitSignal             chan game.EmptySignal
	StopClientSendSignal        chan game.EmptySignal
	SendUserRelatedDataToClient func(clientID string, clientIP *net.IP, clientPort int, stopClientSendSignal chan game.EmptySignal)
	BroadcastUpdateChannel      chan game.EmptySignal
}
