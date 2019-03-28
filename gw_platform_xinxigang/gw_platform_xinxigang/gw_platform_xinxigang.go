// gw_platform_xinxigang project main.go
package main

import (
	"config"
	//"fmt"
	"http_cms"
	//"http_ops"
	"platform"

	"git.scsv.online/go/logger"
)

func main() {

	config.Config_p = new(config.Config)
	err := config.Config_p.Load()
	if err != nil {
		logger.Error("config load error: %v\n", err)
		return
	}

	http_cms.Http_cms_p = new(http_cms.Http_cms)
	err = http_cms.Http_cms_p.Start()
	if err != nil {
		logger.Error("http_cms Start error: %v\n", err)
		return
	}
	//http_ops.Start()

	platform.Platform_p = new(platform.Platform)
	err = platform.Platform_p.Start()
	if err != nil {
		logger.Error("platform Start error: %v\n", err)
		return
	}

	select {}
}
