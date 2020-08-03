package main

import (
	"fmt"
	"time"
)

func main() {
	chanMap := make(map[string]chan struct{})
	chanMap["11"] = make(chan struct{})

	go func() {
		go func() {
			defer fmt.Println("11  func end")
			chanMap["11"] <- struct{}{}
		}()

		fmt.Println("ticker start")
		ticker := time.NewTicker(time.Minute * 2)
		defer ticker.Stop()
		select {
		case <-ticker.C:
			fmt.Println("ticker stop")
			//case <-chanMap["11"]:
			//	fmt.Println("11 stop")
		}
	}()

	go func() {
		//chanMap["11"] <- struct{}{}
		ticker := time.NewTicker(time.Minute * 3)
		defer ticker.Stop()
		select {
		case <-ticker.C:
			fmt.Println("ticker sleep")
		}
	}()

	select {}
}
