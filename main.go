package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
)

const WORKER_COUNT int = 10;
type ChannelAndPort struct {
	Channel chan int
	Port int
}
func main() {
	workerLines := make([]ChannelAndPort, 0, WORKER_COUNT)

	for workerId := 0; workerId < WORKER_COUNT; workerId++  {
		workerLine := make(chan int)
		addr, err := net.ResolveUDPAddr("udp", ":0")

		if err != nil {
			panic(err)
		}
	
		conn, err := net.ListenUDP("udp", addr)
	
		if err != nil {
			panic(err)
		}
	
		port := conn.LocalAddr().(*net.UDPAddr).Port
		
		go func(workerId int, conn *net.UDPConn, workerLine <-chan int) {
			defer conn.Close()

			for message := range workerLine {
				fmt.Println(message)
			}
		}(workerId, conn, workerLine)

		workerLines = append(workerLines, ChannelAndPort{
			Channel: workerLine,
			Port: port,
		})
	}
	// Go에서 protobuf를 사용하기 위해 필요한 단계: https://protobuf.dev/getting-started/gotutorial/
	// ex) protoc --go_out=$PWD proto/status.proto
	http.HandleFunc("/get-worker-port", func(w http.ResponseWriter, r *http.Request) {


		w.Header().Set("Content-Type", "text/plain")
		io.WriteString(w, fmt.Sprintf("%d", port))
	})

	log.Fatal(http.ListenAndServe(":8888", nil))

	// for {
	// 	buffer := make([]byte, 1024)
	// 	amount, _, err := conn.ReadFromUDP(buffer)

	// 	if err != nil {
	// 		log.Fatal(err.Error())
	// 	}

	// 	status := new(protodef.Status)
	// 	desErr := proto.Unmarshal(buffer[:amount], status)

	// 	if desErr != nil {
	// 		log.Fatal(err.Error())
	// 	}
	// }
}
