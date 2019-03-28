package main

import (
	"fmt"
	"net"
	"runtime"
	"time"
)

func client() {
	conn, err := net.Dial("tcp", "192.168.1.132:7509")
	if err != nil {
		fmt.Println("tcp conn err", err)
		return
	}
	defer conn.Close()
	// tConn, ok := conn.(*net.TCPConn)
	// 	if !ok {
	// 		fmt.Println("vsdfgfdghdfhgdfh")
	// 	}
	// 	tConn.SetWriteBuffer(1024 * 1024 * 10)

	buf := make([]byte, 1024)
	for {
		fmt.Println("tcp Write start")
		_, err := conn.Write(buf)
		if err != nil {
			fmt.Println("tcp Write err", err)
			return
		}
		fmt.Printf("向服务端%v发来信息\n", conn.LocalAddr())
		fmt.Println("tcp Write end")
		time.Sleep(time.Duration(10) * time.Second)
	}
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	for i := 0; i < 1000; i++ {
		go client()
		time.Sleep(time.Duration(50) * time.Millisecond)
	}

	select {}
}
