package task

func SendMutualTerminationSignal(signal chan bool) {
	if r := recover(); r != nil {
		signal <- true
	}
	close(signal)
}
