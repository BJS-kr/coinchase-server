package task

import "net"

func MakeTCPListener() *net.TCPListener {
	// 호스트의 아무 빈 포트에 할당한다.
	addr, err := net.ResolveTCPAddr("tcp", ":0")

	if err != nil {
		panic(err)
	}

	tcpListener, err := net.ListenTCP("tcp", addr)

	if err != nil {
		panic(err)
	}

	return tcpListener
}
