package main

import (
	"git.scsv.online/go/base/cache"
	"git.scsv.online/go/base/logger"
)

func main() {
	cache := cache.NewRedis("tcp", "192.168.1.5:33679", 0)
	err := cache.Connect()
	if err != nil {
		logger.Error(err.Error())
		return
	}

	key := "test"

	// 写入
	cache.SetCache(key, "abcd")
	// 读取
	out, err := cache.GetCache(key)
	if err != nil {
		logger.Error(err.Error())
		return
	}

	logger.Debug("%s->%s", key, string(out.([]byte)))

	ch := make(chan bool)

	go cache.ExpireEvent(func(key string) {
		logger.Error(key)
	})

	<-ch
}
