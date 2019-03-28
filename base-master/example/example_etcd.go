package main

import (
	"git.scsv.online/go/base/etcd"
	"git.scsv.online/go/base/logger"
)

//配置文件
type config struct {
	Etcd     string `yaml:"etcd"`
	Type     string `yaml:"type"`
	SN       string `yaml:"sn"`
	Intranet struct {
		IP     string `yaml:"ip"`
		Http   int    `yaml:"http"`
		Rtsp   int    `yaml:"rtsp"`
		Svp    int    `yaml:"svp"`
		Telnet int    `yaml:"telnet"`
		Gnss   int    `yaml:"gnss"`
	} `yaml:"intranet"`
	Extranet struct {
		IP     string `yaml:"ip"`
		Http   int    `yaml:"http"`
		Rtsp   int    `yaml:"rtsp"`
		Svp    int    `yaml:"svp"`
		Telnet int    `yaml:"telnet"`
		Gnss   int    `yaml:"gnss"`
	} `yaml:"extranet"`
	Ops struct {
		Enable   bool   `yaml:"enable"`
		Url      string `yaml:"url"`
		Interval int    `yaml:"interval"`
	} `yaml:"ops"`
	Limit struct {
		CPU    int `yaml:"CPU"`
		Input  int `yaml:"input"`
		Output int `yaml:"output"`
	} `yaml:"limit"`
}

func main() {
	logger.Level = logger.LOG_ALL

	cfg := &config{}
	_, err := etcd.LoadConfig("./config.yml", &cfg, true)
	if err != nil {
		logger.Error(err.Error())
		return
	}

	logger.Info("%+v", cfg)

	ch := make(chan int)
	<-ch

}
