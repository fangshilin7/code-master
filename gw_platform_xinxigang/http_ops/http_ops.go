package http_ops

import (
	"bytes"
	"config"
	"encoding/json"

	"fmt"

	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"git.scsv.online/go/logger"
)

type gateway_info struct {
	Server server `json:"server"`
}

type server struct {
	Version string `json:"version"`
	Build   string `json:"build"`
	Type    string `json:"type"`
}

func Start() {
	logger.Info("http_ops")
	gateway := gateway_info{}
	go gateway.timer_function()
}

func (gateway *gateway_info) timer_function() {

	ticker := time.NewTicker(time.Second * 30)
	defer ticker.Stop()
	for {
		<-ticker.C
		gateway.keep_alive()
	}
}

func (gateway *gateway_info) keep_alive() {

	gateway.Server.Version = "1.0"
	gateway.Server.Build = "2019-02-22 10:22:45"
	gateway.Server.Type = config.Config_p.Type

	body, _ := json.Marshal(gateway)
	logger.Info("keep_alive:[%s]", string(body))

	defer func() {
		if err := recover(); err != nil {
			log.Fatalf("keep_alive error:%v", err)
			return
		}
	}()
	rep_body, _ := gateway.http_report(string(body))
	logger.Info("keep_alive resp:%v", rep_body)
}

func (gateway *gateway_info) http_report(body string) (string, error) {
	var url bytes.Buffer
	fmt.Fprintf(&url, "%s/status/%s/%s", config.Config_p.Ops, config.Config_p.Type, config.Config_p.Gateway)
	logger.Info("http_report url:[%s]", url.String())
	resp, err := http.Post(url.String(), "application/x-www-form-urlencoded", strings.NewReader(body))
	defer resp.Body.Close()

	rep_body, err := ioutil.ReadAll(resp.Body)
	return string(rep_body), err
}
