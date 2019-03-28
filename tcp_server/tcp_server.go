package main

import (
	"fmt"
	"net"
	"runtime"
)

//var simu_buf []byte = make([]byte, 1024*160)

func Handle(conn net.Conn) {
	defer conn.Close()
	buf := make([]byte, 1024*160)

	for {
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Println("read err", err)
			return
		}
		//messsage := buf[:n]
		fmt.Println("客户器端发来信息", conn.RemoteAddr(), n)
	}
}
func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	listen, err := net.Listen("tcp", "192.168.1.132:7509")
	if err != nil {
		fmt.Println("listen err", err)
		return
	}
	defer listen.Close()
	for {
		conn, err := listen.Accept()
		if err != nil {
			fmt.Println("accept err", err)
			return
		}
		go Handle(conn)
	}
}
