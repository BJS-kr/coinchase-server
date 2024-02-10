package task

import "net"

func MakeUDPConn() *net.UDPConn {
	// 호스트의 아무 빈 포트에 할당한다.
	addr, err := net.ResolveUDPAddr("udp", ":0")

	if err != nil {
		panic(err)
	}

	conn, err := net.ListenUDP("udp", addr)

	if err != nil {
		panic(err)
	}

	return conn
}
