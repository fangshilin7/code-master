# Golang Base Library

## 安装
需要Go 1.8以上支持，[点击下载](https://golang.org/)

## Usage

### logger
```go
package main

import (
	"git.scsv.online/go/base/logger"
)

func main() {
	logger.Debug("Hello GO, %d %s", 1, "!")
	logger.Info("Hello GO")
	logger.Warn("Hello GO")
	logger.Error("Hello GO")
	return
}
```

### rtsp

#### client
```go
package main

import (
	"fmt"
	"sync"
	"time"
	"git.scsv.online/go/base/rtsp"
	"git.scsv.online/go/base/logger"
)


var packetPool = sync.Pool{
	New: func() interface{}{
		p := &rtsp.RtspPacket{}
		return p
	},
}

type DataReceiver struct {
}

func (*DataReceiver) OnError(packet *rtsp.RtspPacket, err error) {
	fmt.Println("OnError :", err)
	packetPool.Put(packet)
}

func (*DataReceiver) OnPacket(packet *rtsp.RtspPacket) (err error) {
	if packet.Msg != nil {
		fmt.Printf("ReceiveMessage : %s", packet.Msg.Raw)
	}
	
	if packet.Stream != nil {
		fmt.Printf("PacketStream : %d", len(packet.Stream))
	}
	
	packetPool.Put(packet)
	return nil
}

func (*DataReceiver) PacketPool() (*sync.Pool){
	return &packetPool
}

func main() {
	const RTSPURL = "rtsp://admin:12345@192.168.1.62:554/Streaming/Channels/101?transportmode=unicast&profile=Profile_101"
    var receiver DataReceiver
    ch := make(chan int)
    s, err := rtsp.Connect(RTSPURL, &receiver, ch)
    if err != nil {
        logger.Error(err.Error())
        return
    }


    <-ch
    logger.Debug("Stream Ready")
    
    <- time.After(time.Second * 60)
    logger.Debug("Time Closed")
    s.Close()

	return
}

```

#### server

```go
package main

import (
	"runtime"
	"git.scsv.online/go/base/rtsp"
	"git.scsv.online/go/base/logger"
)

func main(){
	runtime.GOMAXPROCS(runtime.NumCPU()*2)
    srv, err := rtsp.RunServer(8001)
    if err != nil {
        logger.Error(err.Error())
    }else{
        <- srv.Exit
    }
    
    logger.Info("Program Exited.")
}

```