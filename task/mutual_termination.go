package task

func SendMutualTerminationSignal(terminationSignal chan bool) {
	if r := recover(); r != nil {
		// close는 이 채널을 구독하는 모든 goroutine에게 zero value를 보내므로 단순히 close하는 것만으로도 모든 goroutine을 정리할 수 있다.
		close(terminationSignal)
	}
}
