// 带宽测试工具
// 服务端： nb 192.168.3.6:50000 server
// 客户端： nb 192.168.3.6:50000

package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"runtime"
	"time"
)

const GB = 1 << 30
const MB = 1 << 20
const KB = 1 << 10

var transBytes int64

func usage() {
	fmt.Println(`usage: nb ip:port [server]`)
}

func server(ln net.Listener) {
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Print(err)
			break
		}
		log.Printf("Client Connectd: %s", conn.RemoteAddr().String())
		go func(c net.Conn) {
			buf := make([]byte, 1024*1024)
			for {
				n, err := c.Write(buf)
				if err != nil {
					break
				}
				transBytes += int64(n)
			}
			log.Printf("Client Closed: %s", c.RemoteAddr().String())
		}(conn)
	}
}

func client(conn net.Conn) {
	buf := make([]byte, 1024*1024)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			break
		}
		transBytes += int64(n)
	}
	log.Println("Client Closed")
}

func main() {
	if len(os.Args) < 2 {
		usage()
		return
	}

	runtime.GOMAXPROCS(1)
	host := os.Args[1]
	if len(os.Args) > 2 {
		//server
		ln, err := net.Listen("tcp", host)
		if err != nil {
			log.Printf(err.Error())
			return
		}
		log.Printf("Server Listen, %s", host)
		go server(ln)
	} else {
		conn, err := net.Dial("tcp", host)
		if err != nil {
			log.Println(err)
			return
		}
		log.Printf("Connected, %s", host)
		go client(conn)
	}

	go func() {
		for {
			select {
			case <-time.After(time.Second * 5):
				{
					bytes := (transBytes * 8) / 5
					if bytes > GB {
						log.Printf("%d Gbps", bytes/GB)
					} else if bytes > MB {
						log.Printf("%d Mbps", bytes/MB)
					} else {
						log.Printf("%d Kbps", bytes/KB)
					}

					transBytes = 0
				}
			}
		}
	}()

	ch := make(chan int)
	<-ch
}
