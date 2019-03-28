// gw_platform_xinxigang project doc.go

/*
gw_platform_xinxigang document
*/
package http_cms

import (
	"bytes"
	"config"
	"encoding/json"
	"strings"

	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"

	//"strings"
	"sync"
	"time"

	"git.scsv.online/go/logger"
)

var Http_cms_p *Http_cms

type resp struct {
	Data struct {
		Devices []Device
	}
	Status struct {
		Code    int
		Message string
	}
}

type Device struct {
	Id      string
	OuterId string
	ResType int
	Name    string
	Ip      string
}

type Http_cms struct {
	Dev_mutex    sync.Mutex
	Devices_info resp
}

/*
func new_Http_cms() *Http_cms {
	return &Http_cms{
		Report_cmd_vld: "trafData",
		Report_cmd_mdi: "meteData",
		Report_cmd_vtd: "trafEvent",
	}
}
*/

func (http_cms *Http_cms) Start() error {

	//Http_report_cmd = new_Http_cms()

	http_cms.request_devices()

	return nil
}

func (http_cms *Http_cms) http_request(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	data, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return "", err
	}
	//logger.Info("http resp: %s", data)
	return string(data), nil
}

func (http_cms *Http_cms) request_devices() {
	var url bytes.Buffer
	fmt.Fprintf(&url, "%s/gwDevice?id=%s&res=%d&rc=%d", config.Config_p.Cms, config.Config_p.Gateway, 10, http_cms.random_rc())
	logger.Info("request_devices url:[%s]", url.String())

	defer func() {
		if err := recover(); err != nil {
			logger.Error("http_request error:%v", err)
			return
		}
	}()
	data, _ := http_cms.http_request(url.String())
	logger.Info("http_request resp:[%s]", data)

	err := json.Unmarshal([]byte(data), &http_cms.Devices_info)
	if err != nil {
		logger.Error("json Unmarshal error:%v", err)
		return
	}
	logger.Info("request_devices code:[%d], Message:[%s]", http_cms.Devices_info.Status.Code, http_cms.Devices_info.Status.Message)
}

func (http_cms *Http_cms) random_rc() int {
	rand.Seed(time.Now().Unix())
	rnd := rand.Int()
	return rnd
}

func (http_cms *Http_cms) Report_device_status(body []byte, rep_cmd string) {
	defer func() {
		if err := recover(); err != nil {
			logger.Error("Report_device_status error:%v", err)
			return
		}
	}()

	//logger.Info("Report_device_status body:%s", string(body))
	var url bytes.Buffer
	fmt.Fprintf(&url, "%s/%s?rc=%d", config.Config_p.Cms, rep_cmd, http_cms.random_rc())
	logger.Info("report url:[%s]", url.String())

	req, _ := http.NewRequest("POST", url.String(), strings.NewReader(string(body)))
	req.Header.Add("Content-Type", "application/json")
	//req.Header.Add("cache-control", "no-cache")
	//req.Header.Add("Postman-Token", "f572bcdb-b98a-41d3-a2b4-50e3544bc242")

	resp, _ := http.DefaultClient.Do(req)

	//resp, _ := http.Post(url.String(), "application/x-www-form-urlencoded", bytes.NewReader(body)) //strings.NewReader(body))

	defer resp.Body.Close()
	rep_body, _ := ioutil.ReadAll(resp.Body)
	logger.Info("Report_device_status resp:%s", string(rep_body))
}

func (http_cms *Http_cms) Find_valid_device(outer_id string) (string, bool) {
	//http_cms.Dev_mutex.Lock()
	//defer
	for _, dev := range http_cms.Devices_info.Data.Devices {
		if dev.OuterId == outer_id {
			return dev.Id, true
		}
	}
	//http_cms.Dev_mutex.Unlock()
	return "", false
}
