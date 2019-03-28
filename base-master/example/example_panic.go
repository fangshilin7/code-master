package main

import (
	"git.scsv.online/go/base/logger"
	"git.scsv.online/go/base/util"
	"time"
)

func fatalFunc() {
	var bytes []byte
	bytes[100] = '1'
}

func testPanic() {
	defer util.PanicTrace(true)
	fatalFunc()
	logger.Debug("123")
}

func main() {
	defer util.PanicTrace(false)
	go testPanic()
	for i := 0; i < 5; i++ {
		time.Sleep(time.Second)
		logger.Debug("ok")
		if i == 4 {
			a := 0
			a = 1 / a
		}
	}
	//fatalFunc()
	//ch := make(chan int)
	//<-ch
}
