package util

import (
	"fmt"
	"git.scsv.online/go/base/logger"
	"runtime/debug"
)

func PanicTrace(exit bool) {
	if err := recover(); err != nil {
		dumpstr := fmt.Sprintf("\r\n%+v\r\n", err)

		//s := []byte("/src/runtime/panic.go")
		//e := []byte("\ngoroutine ")
		//line := []byte("\n")
		//stack := make([]byte, 4096) //4KB
		//length := runtime.Stack(stack, true)
		//start := bytes.Index(stack, s)
		//stack = stack[start:length]
		//start = bytes.Index(stack, line) + 1
		//stack = stack[start:]
		//end := bytes.LastIndex(stack, line)
		//if end != -1 {
		//	stack = stack[:end]
		//}
		//end = bytes.Index(stack, e)
		//if end != -1 {
		//	stack = stack[:end]
		//}
		//stack = bytes.TrimRight(stack, "\n")

		dumpstr += fmt.Sprintf("%s\r\n", debug.Stack())

		logger.Fatal(dumpstr)
		if exit {
			panic(err)
		}
	}
}
