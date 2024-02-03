package main

import (
	"fmt"
	"log"
	"mrpg/protodef"
	"net"

	"google.golang.org/protobuf/proto"
)

func main() {
	// Go에서 protobuf를 사용하기 위해 필요한 단계: https://protobuf.dev/getting-started/gotutorial/
	// ex) protoc --go_out=$PWD proto/status.proto  
	addr, err := net.ResolveUDPAddr("udp", ":8888")

	if err != nil {
		panic(err)
	}

	conn, err := net.ListenUDP("udp", addr)

	if err != nil {
		panic(err)
	}

	defer conn.Close()

	for {
		buffer := make([]byte, 1024)
		amount, _, err := conn.ReadFromUDP(buffer)
	
		if err != nil {
			log.Fatal(err.Error())
		}

		status := new(protodef.Status)
		desErr := proto.Unmarshal(buffer[:amount], status)

		if desErr != nil {
			log.Fatal(err.Error())
		}

		fmt.Println(*status)
	}
}