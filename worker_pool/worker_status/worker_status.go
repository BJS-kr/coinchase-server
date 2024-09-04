package worker_status

const (
	IDLE = iota + 1
	PULLED_OUT
	CLIENT_INFORMATION_RECEIVED
	WORKING
	TERMINATED
)
