package platform

import (

	//"bytes"
	"errors"
	//"fmt"
	"config"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"sort"
	"strconv"

	"http_cms"
	"strings"
	"time"

	"git.scsv.online/go/logger"
	//	"github.com/tidwall/gjson"
)

const nodata_time_max = 3 //平台请求错误最大次数
const dev_type_count = 3  //当前设备类型数量
const vld_type = 0        //解析数据时用于区分设备类型的命令
const mdi_type = 1
const vtd_type = 2

const resType_vld = 12 //中心设备类型
const resType_mdi = 13
const resType_vtd = 14
const resType_rvm = 15

type Platform struct {
	Http_infos [dev_type_count]http_info
}

type http_info struct {
	Url          string
	Token        string
	Body         string
	Rep_cmd      string
	nodata_times int
	base_time    string
	client       *http.Client
	Devices      []http_cms.Device
}

var Platform_p *Platform

/*
type report_exception struct {
	Gateway        string `json:"gateWay"`
	Devicestatus   bool   `json:"deviceStatus"`
	Internetstatus bool   `json:"internetStatus"`
}*/

type report_str struct {
	Gateway        string        `json:"gateWay"`
	Data           []interface{} `json:"data"`
	Exception      []string      `json:"exception"`
	Devicestatus   bool          `json:"deviceStatus"`
	Internetstatus bool          `json:"internetStatus"`
}

type vld_str struct {
	Outerid       string  `json:"outerID"`
	Name          string  `json:"name"`
	Totalflow     int     `json:"totalFlow"`
	Vehiclebig    int     `json:"vehicleBig"`
	Vehiclemiddle int     `json:"vehicleMiddle"`
	Vehiclesmall  int     `json:"vehicleSmall"`
	Avgspeed      float32 `json:"avgSpeed"`
	Avgoccrate    int     `json:"avgOccRate"`
}

type mdi_str struct {
	Outerid                string  `json:"outerID"`
	Coltime                string  `json:"colTime"`
	Inswinddir             float32 `json:"insWindDir"`
	Inswindspeed           float32 `json:"insWindSpeed"`
	Humidity               float32 `json:"humidity"`
	Airpressure            float32 `json:"airPressure"`
	Temperature            float32 `json:"temperature"`
	Roadbedtemperature     float32 `json:"roadbedTemperature"`
	Visibility             float32 `json:"visibility"`
	Roadsurfacetemperature float32 `json:"roadSurfaceTemperature"`
}

type vtd_str struct {
	Event_type      string `json:"type"`
	Version         string `json:"outerID"`
	Datasource      int    `json:"dataSource"`
	Outerid         string `json:"outerID"`
	Devicelocation  string `json:"deviceLocation"`
	Eventcreatetime string `json:"eventCreateTime"`
	Eventstarttime  string `json:"eventStartTime"`
	Eventendtime    string `json:"eventEndTime"`
	Lanenumber      int    `json:"laneNumber"`
	Eventtype       string `json:"eventType"`
	Drection        string `json:"direction"`
	Eventlocationx  int    `json:"eventLocationX"`
	Eventlocationy  int    `json:"eventLocationY"`
	Picnumber       int    `json:"picNumber"`
	Pics            []pic  `json:"pics"`
	Startvideotime  string `json:"startVideoTime"`
	Endvideotime    string `json:"endVideoTime"`
	Videourl        string `json:"videoURL"`
}

type vtd_time struct {
	Starttime string `json:"starttime"`
	Endtime   string `json:"endtime"`
}

type pic struct {
	Url  string `json:"url"`
	Name string `json:"name"`
}

type response_vld_str struct {
	Status  bool
	Message string
	Data    []string
}

type response_mdi_str struct {
	Status  bool
	Message string
	Data    []struct {
		SID  string
		TT   string
		AA   string
		AB   string
		BD   string
		BP   string
		BC   string
		APB  string
		HB1T string
		NL   string
	}
}

type response_vtd_str struct {
	Datasource     string
	Deviceid       string
	Devicepos      string
	Laneno         string
	Eventtype      string
	Eventtime      string
	Starttime      string
	Endtime        string
	Videostarttime string
	Videoendtime   string
	Direction      string
	Piccount       string
	Picx           string
	Picy           string
	Inserttime     string
	Pic1           string
	Pic2           string
	Videopath      string
}

func (platform_info *Platform) Start() error {
	logger.Info("platform")
	//httpInfo.Http_infos.append(config.Config_p.Platform.Vld.Url, config.Config_p.Platform.Vld.Token, "{}")
	platform_info.Http_infos[vld_type] = http_info{config.Config_p.Platform.Vld.Url, config.Config_p.Platform.Vld.Token, "{}", "trafData", 0, "", &http.Client{Timeout: 10 * time.Second}, http_cms.Http_cms_p.Devices_info.Data.Devices}
	platform_info.Http_infos[mdi_type] = http_info{config.Config_p.Platform.Mdi.Url, config.Config_p.Platform.Mdi.Token, "{}", "meteData", 0, "", &http.Client{Timeout: 10 * time.Second}, http_cms.Http_cms_p.Devices_info.Data.Devices}
	platform_info.Http_infos[vtd_type] = http_info{config.Config_p.Platform.Vtd.Url, config.Config_p.Platform.Vtd.Token, "{}", "trafEvent", 0, "", &http.Client{Timeout: 10 * time.Second}, http_cms.Http_cms_p.Devices_info.Data.Devices}

	for i, dev_type := range platform_info.Http_infos {
		//if i == vld_type {
		go dev_type.timer_function(i, dev_type.Url, dev_type.Token, dev_type.Body, dev_type.Rep_cmd)
		//}
	}
	return nil
}

//假数据测试
func (platform_info *Platform) test() {

}

func (http_p *http_info) timer_function(cmd int, url, token, body, rep_cmd string) {
	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			logger.Info("timer_function:%d", cmd)
			http_p.get_state(cmd, url, token, body, rep_cmd)
		}
	}
}

// 定时获取平台信息
func (http_p *http_info) get_state(cmd int, url, token, body, rep_cmd string) {
	logger.Info("get_state:%d start", cmd)
	defer func() {
		if err := recover(); err != nil {
			logger.Error("get_state %d error:[%v]", cmd, err)
			return
		}
	}()
	/*
		defer func() {
			logger.Info("get_state After 5 Second")
			<-time.After(time.Second * time.Duration(5))
			platform_info.get_state(cmd, url, token, body)
		}()*/

	//事件查询方式为时间段查询，构建时间段请求体
	if cmd == vtd_type {
		//body =
		cur_time := time.Now().Format("2006/01/02 15:04:05")
		if http_p.base_time == "" {
			http_p.base_time = cur_time
		}
		var time_info vtd_time
		time_info.Starttime = http_p.base_time
		time_info.Endtime = cur_time
		http_p.base_time = cur_time

		time_body, _ := json.Marshal(time_info)
		body = string(time_body)
		logger.Info("time_body: %s", body)
	}
	//var resp_data string
	resp_data, err := http_p.http_request(url, token, body)
	if err != nil {

		logger.Warn("http_request url:%s failed:[[%v]]", url, err)
		if cmd == vtd_type {
			//return
		}
		http_p.nodata_times++
		if http_p.nodata_times >= nodata_time_max {
			//构建设备异常数据
			rep_str := report_str{}
			rep_str.Gateway = config.Config_p.Gateway
			rep_str.Devicestatus = false
			rep_str.Internetstatus = false

			//body, _ := json.Marshal(rep_str)

			//go http_cms.Http_cms_p.Report_device_status(body, http_p.Rep_cmd)
			http_p.nodata_times = 0
		}
		//return
		//{
		//假数据
		if cmd == mdi_type {
			resp_data = `{
    "status" : true, 
    "message" : "SUCCESS", 
    "data" : [
	{
		"SID" : "84446654",
		"TT" : "2018-12-18 15:22:10",
		"AA" : "339.0",
		"AB" : "11.0",
		"BD" : "70.0",
		"BP" : "9556.0",
		"BC" : "195.0",
		"APB" : "198",
		"HB1T" : "6292",
		"NL" : "228",
		"TEST" : ""
	},
	{
		"SID" : "C0001",
		"TT" : "2018-12-18 15:22:10",
		"AA" : "350.0",
		"AB" : "13.0",
		"BD" : "72.0",
		"BP" : "9550.0",
		"BC" : "196.0",
		"APB" : "190",
		"HB1T" : "6290",
		"NL" : "220",
		"TEST" : ""
	},
	{
		"SID" : "C0002",
		"TT" : "2018-12-18 15:22:10",
		"AA" : "333.0",
		"AB" : "10.0",
		"BD" : "74.0",
		"BP" : "9500.0",
		"BC" : "190.0",
		"APB" : "180",
		"HB1T" : "6200",
		"NL" : "200",
		"TEST" : ""
	}]
}`
		} else if cmd == vld_type {
			resp_data = `{
    "status" : true, 
    "message" : "SUCCESS", 
    "data" : [
		"VOLUME,1.0,3014-321000,510100000000A90080,20.3.134.16,入高速新繁收费站分流点2入高速,2019-02-21 17:40:00,60,2,2,2B,2,0,0,2,0,0,37.0,0,1,200.0,2019-02-22 01:38:16",
        "VOLUME,1.0,3014-321001,510100000000A90077,20.3.134.13,成彭高速公路（川高速S1）13公里228米（彭州至成都）,2019-02-21 17:40:00,60,3,5,2B,1,0,0,1,0,0,45.0,0,1,200.0,2019-02-22 01:38:16",
        "VOLUME,1.0,3014-321025,510100000000A32096,20.3.133.9,成彭高速公路（川高速S1）6公里919米（彭州至成都）,2019-02-21 17:40:00,60,3,5,2A,0,0,0,0,0,0,0.0,0,49,200.0,2019-02-21 17:34:37",
        "VOLUME,1.0,3014-321026,510100000000A90034,20.3.132.38,绕城外圈K0+850分流点外圈,2019-02-21 17:40:00,60,3,2,2A,20,0,0,20,0,0,19.0,0,9,13.0,2019-02-21 17:34:37",
        "VOLUME,1.0,3014-321027,510100000000A90071,20.3.134.7,成彭高速公路（川高速S1）12公里887米（成都至彭州）,2019-02-21 17:40:00,60,3,3,2B,2,0,0,2,0,0,0.0,0,1,0.0,2019-02-22 01:38:16",
        "VOLUME,1.0,3014-321028,,20.3.134.55,成彭高速出成都14km+650m处,2019-02-21 17:40:00,60,6,1,,0,0,0,0,0,0,0.0,0,2,200.0,2019-02-22 01:37:04",
        "VOLUME,1.0,3014-321031,510100000000A32094,20.3.133.8,成彭高速出城6km+700m处,2019-02-21 17:40:00,60,3,3,2B,3,0,0,3,0,0,82.0,0,1,200.0,2019-02-21 17:34:37",
        "VOLUME,1.0,3014-321032,510100000000A90082,20.3.134.18,成彭高速公路(川高速S1)13公里582米(成都至彭州),2019-02-21 17:40:00,60,3,4,2B,1,0,0,1,0,0,0.0,0,1,0.0,2019-02-22 01:38:16",
        "VOLUME,1.0,3014-321033,,20.3.133.39,成彭高速公路K5+200,2019-02-21 17:40:00,60,8,4,2A,0,0,0,0,0,0,0.0,0,0,200.0,2019-02-22 01:37:04",
        "VOLUME,1.0,3014-321034,510100000000A90058,20.3.133.18,入高速龙桥收费站分流点入高速,2019-02-21 17:40:00,60,4,4,2B,3,0,0,3,0,0,27.0,12,50,200.0,2019-02-22 01:38:16",
        "VOLUME,1.0,3014-321035,,20.3.135.7,成彭高速进成都19km+700m处,2019-02-21 17:40:00,60,3,2,2A,5,0,0,5,0,0,57.0,0,2,183.0,2019-02-21 17:34:37",
        "VOLUME,1.0,3014-321036,,20.3.132.79,成彭高速出成都1km+380m处,2019-02-21 17:40:00,60,8,-1,,6,0,0,6,0,0,5.0,0,2,16.0,2019-02-22 01:37:04",
        "VOLUME,1.0,3014-321037,510100000000A90057,20.3.133.17,入高速龙桥收费站分流点入高速,2019-02-21 17:40:00,60,2,2,,15,0,0,15,0,0,31.0,0,8,29.0,2019-02-22 01:38:16",
        "VOLUME,1.0,3014-321038,510100000000A90039,20.3.132.43,绕城内圈K1+250合流点内圈,2019-02-21 17:40:00,60,3,3,,4,0,0,4,0,0,0.0,0,3,0.0,2019-02-21 17:34:38",
        "VOLUME,1.0,3014-321039,510100000000A9EF4X,20.1.26.52,新成彭路(洞子口路口至三环路口),2019-02-21 17:40:00,60,3,3,2B,2,0,1,1,0,0,19.0,0,2,126.0,2019-02-21 17:40:04",
        "VOLUME,1.0,3014-321042,510100000000A9GAT9,20.1.15.38,三环路成彭立交匝道(三环路凤凰立交至成彭路),2019-02-21 17:40:00,60,2,2,1B,8,0,0,8,0,0,54.6,3,3,113.8,2019-02-21 17:41:00",
        "VOLUME,1.0,3014-321043,510100000000A90088,20.3.134.26,成彭高速公路（川高速S1）15公里890米（彭州至成都）,2019-02-21 17:41:00,60,2,4,2B,3,0,0,3,0,0,44.0,0,1,200.0,2019-02-22 01:39:16",
        "VOLUME,1.0,3014-321046,510100000000A90060,20.3.133.20,成彭高速公路(川高速S1)8公路120米(彭州至成都),2019-02-21 17:41:00,60,2,5,2B,4,0,0,4,0,0,25.0,0,1,135.0,2019-02-22 01:39:16",
        "VOLUME,1.0,3014-321048,510100000000A90051,20.3.133.11,成彭高速公路（川高速S1）7公里470米（成都至彭州）,2019-02-21 17:41:00,60,3,5,2A,6,0,3,3,0,0,46.0,0,2,110.0,2019-02-22 01:39:16",
        "VOLUME,1.0,3014-321049,510100000000A90057,20.3.133.17,入高速龙桥收费站分流点入高速,2019-02-21 17:41:00,60,2,2,,7,0,0,7,0,0,28.0,0,4,48.0,2019-02-22 01:39:16",
        "VOLUME,1.0,3014-321050,510100000000A90034,20.3.132.38,绕城外圈K0+850分流点外圈,2019-02-21 17:41:00,60,3,1,2A,4,0,0,4,0,0,36.0,0,2,200.0,2019-02-21 17:35:37",
        "VOLUME,1.0,3014-321051,510100000000A90050,20.3.133.6,成彭高速公路（川高速S1）5公里367米（成都至彭州）,2019-02-21 17:41:00,60,2,5,2B,0,0,0,0,0,0,0.0,0,0,200.0,2019-02-22 01:39:16",
        "VOLUME,1.0,3014-321052,510100000000A90058,20.3.133.18,入高速龙桥收费站分流点入高速,2019-02-21 17:41:00,60,4,3,2B,9,0,0,9,0,0,32.0,0,3,58.0,2019-02-22 01:39:16",
        "VOLUME,1.0,3014-321053,510100000000A90089,20.3.134.27,成彭高速公路（川高速S1）15公里889米（成都至彭州）,2019-02-21 17:41:00,60,3,3,2B,2,0,2,0,0,0,80.0,0,2,200.0,2019-02-22 01:39:16",
        "VOLUME,1.0,3014-321054,,20.3.135.7,成彭高速进成都19km+700m处,2019-02-21 17:41:00,60,3,2,2A,9,0,0,9,0,0,64.0,0,4,106.0,2019-02-21 17:35:37",
        "VOLUME,1.0,3014-321055,,20.3.132.79,成彭高速出成都1km+380m处,2019-02-21 17:41:00,60,8,4,,0,0,0,0,0,0,0.0,5,50,200.0,2019-02-22 01:38:04",
        "VOLUME,1.0,3014-321056,510100000000A90069,20.3.134.5,成彭高速公路（川高速S1）11公里695米（彭州至成都）,2019-02-21 17:41:00,60,2,5,2B,0,0,0,0,0,0,0.0,0,0,200.0,2019-02-22 01:39:16",
        "VOLUME,1.0,3014-321057,510100000000A90061,20.3.133.21,成彭高速公路（川高速S1）8公里205米（成都至彭州）,2019-02-21 17:41:00,60,3,6,2B,0,0,0,0,0,0,0.0,0,50,200.0,2019-02-22 01:39:16",
        "VOLUME,1.0,3014-321058,510100000000A90082,20.3.134.18,成彭高速公路(川高速S1)13公里582米(成都至彭州),2019-02-21 17:41:00,60,3,6,2B,0,0,0,0,0,0,0.0,0,0,200.0,2019-02-22 01:39:16",
        "VOLUME,1.0,3014-321059,510100000000A90090,20.3.134.28,成彭高速公路（川高速S1）15公里889米（成都至彭州）,2019-02-21 17:41:00,60,2,5,2B,3,0,0,3,0,0,0.0,0,1,0.0,2019-02-22 01:39:16",
        "VOLUME,1.0,3014-321060,510100000000A32096,20.3.133.9,成彭高速公路（川高速S1）6公里919米（彭州至成都）,2019-02-21 17:41:00,60,3,3,2A,11,0,10,1,0,0,82.0,0,3,111.0,2019-02-21 17:35:38",
        "VOLUME,1.0,3014-321061,,20.3.134.55,成彭高速出成都14km+650m处,2019-02-21 17:41:00,60,6,3,,0,0,0,0,0,0,0.0,21,50,200.0,2019-02-22 01:38:05",
        "VOLUME,1.0,3014-321062,510100000000A90073,20.3.134.9,成彭高速公路(川高速S1)12公路870米(彭州至成都),2019-02-21 17:41:00,60,3,3,2A,6,0,0,6,0,0,0.0,0,3,0.0,2019-02-22 01:39:17",
        "VOLUME,1.0,3014-321069,,20.3.132.79,成彭高速出成都1km+380m处,2019-02-21 17:42:00,60,8,4,,0,0,0,0,0,0,0.0,5,50,200.0,2019-02-22 01:39:04",
        "VOLUME,1.0,3014-321071,510100000000A90077,20.3.134.13,成彭高速公路（川高速S1）13公里228米（彭州至成都）,2019-02-21 17:42:00,60,3,4,2B,0,0,0,0,0,0,0.0,0,0,200.0,2019-02-22 01:40:16",
        "VOLUME,1.0,3014-321072,510100000000A90057,20.3.133.17,入高速龙桥收费站分流点入高速,2019-02-21 17:42:00,60,2,1,,0,0,0,0,0,0,0.0,0,0,200.0,2019-02-22 01:40:16",
        "VOLUME,1.0,3014-321073,510100000000A90089,20.3.134.27,成彭高速公路（川高速S1）15公里889米（成都至彭州）,2019-02-21 17:42:00,60,3,2,2B,13,0,1,12,0,0,94.0,0,3,104.0,2019-02-22 01:40:16",
        "VOLUME,1.0,3014-321074,510100000000A90079,20.3.134.15,出高速新繁收费站合流点1出高速,2019-02-21 17:42:00,60,2,3,2B,0,0,0,0,0,0,0.0,0,0,200.0,2019-02-22 01:40:16",
        "VOLUME,1.0,3014-321075,,20.3.134.55,成彭高速出成都14km+650m处,2019-02-21 17:42:00,60,6,2,,0,0,0,0,0,0,0.0,0,1,200.0,2019-02-22 01:39:04",
        "VOLUME,1.0,3014-321076,510100000000A90034,20.3.132.38,绕城外圈K0+850分流点外圈,2019-02-21 17:42:00,60,3,1,2A,8,0,0,8,0,0,32.0,0,3,43.0,2019-02-21 17:36:37",
        "VOLUME,1.0,3014-321077,510100000000A90056,20.3.133.16,出高速龙桥互通匝道2出高速,2019-02-21 17:42:00,60,1,2,,6,0,0,6,0,0,30.0,0,2,152.0,2019-02-22 01:40:16",
        "VOLUME,1.0,3014-321078,,20.3.133.39,成彭高速公路K5+200,2019-02-21 17:42:00,60,8,-2,2A,0,0,0,0,0,0,0.0,0,0,200.0,2019-02-22 01:39:04",
        "VOLUME,1.0,3014-321079,510100000000A90093,20.3.134.35,成彭高速公路(川高速S1)17公里400米(彭州至成都),2019-02-21 17:42:00,60,2,1,2A,7,0,0,7,0,0,0.0,0,2,0.0,2019-02-22 01:40:16",
        "VOLUME,1.0,3014-321080,510100000000A32089,20.3.132.127,成彭高速公路(川高速S1)2公里+900米(彭州至成都),2019-02-21 17:42:00,60,2,1,2A,20,0,0,20,0,0,45.0,0,7,33.0,2019-02-21 17:36:37",
        "VOLUME,1.0,3014-321081,510100000000A32107,20.3.134.31,成彭高速公路（川高速S1）16公里728米（成都至彭州）,2019-02-21 17:42:00,60,2,1,2B,7,0,0,7,0,0,104.0,0,2,200.0,2019-02-21 17:36:37",
        "VOLUME,1.0,3014-321082,510100000000A90071,20.3.134.7,成彭高速公路（川高速S1）12公里887米（成都至彭州）,2019-02-21 17:42:00,60,3,2,2B,4,0,0,4,0,0,0.0,0,1,0.0,2019-02-22 01:40:17",
        "VOLUME,1.0,3014-321083,510100000000A90051,20.3.133.11,成彭高速公路（川高速S1）7公里470米（成都至彭州）,2019-02-21 17:43:00,60,3,4,2A,0,0,0,0,0,0,0.0,14,0,200.0,2019-02-22 01:41:16",
        "VOLUME,1.0,3014-321084,510100000000A32096,20.3.133.9,成彭高速公路（川高速S1）6公里919米（彭州至成都）,2019-02-21 17:43:00,60,3,4,2A,4,0,4,0,0,0,85.0,0,2,200.0,2019-02-21 17:37:37",
        "VOLUME,1.0,3014-321085,510100000000A90104,20.3.135.10,成都至彭州K20+720高速出,2019-02-21 17:43:00,60,2,3,2A,0,0,0,0,0,0,0.0,0,0,200.0,2019-02-22 01:41:16",
        "VOLUME,1.0,3014-321086,510100000000A90080,20.3.134.16,入高速新繁收费站分流点2入高速,2019-02-21 17:43:00,60,2,2,2B,1,0,0,1,0,0,35.0,0,1,200.0,2019-02-22 01:41:16",
        "VOLUME,1.0,3014-321087,,20.3.132.79,成彭高速出成都1km+380m处,2019-02-21 17:43:00,60,8,-2,,2,0,0,2,0,0,15.0,0,4,145.0,2019-02-22 01:40:04",
        "VOLUME,1.0,3014-321088,510100000000A90073,20.3.134.9,成彭高速公路(川高速S1)12公路870米(彭州至成都),2019-02-21 17:43:00,60,3,2,2A,12,0,0,12,0,0,0.0,0,5,0.0,2019-02-22 01:41:16",
        "VOLUME,1.0,3014-321089,510100000000A9EF4Y,20.1.26.53,金丰高架桥上(沙河源南路口至三环路成彭立交),2019-02-21 17:43:00,60,3,3,2B,2,0,2,0,0,0,49.0,0,1,200.0,2019-02-21 17:43:01",
        "VOLUME,1.0,3014-321090,510100000000A32108,20.3.134.32,成彭高速公路（川高速S1）16公里728米（成都至彭州）,2019-02-21 17:43:00,60,3,5,2B,0,0,0,0,0,0,0.0,0,1,200.0,2019-02-21 17:37:37",
        "VOLUME,1.0,3014-321091,,20.3.133.39,成彭高速公路K5+200,2019-02-21 17:43:00,60,8,1,2A,0,0,0,0,0,0,0.0,0,0,200.0,2019-02-22 01:40:04",
        "VOLUME,1.0,3014-321095,510100000000A90094,20.3.134.36,成彭高速公路（川高速S1）17公里400米（彭州至成都）,2019-02-21 17:43:00,60,3,3,2A,11,0,0,11,0,0,0.0,0,5,0.0,2019-02-22 01:41:16",
        "VOLUME,1.0,3014-321096,,20.3.135.7,成彭高速进成都19km+700m处,2019-02-21 17:43:00,60,3,2,2A,13,0,0,13,0,0,48.0,0,9,47.0,2019-02-21 17:37:37",
        "VOLUME,1.0,3014-321097,510100000000A90039,20.3.132.43,绕城内圈K1+250合流点内圈,2019-02-21 17:43:00,60,3,2,,6,0,0,6,0,0,0.0,0,3,0.0,2019-02-21 17:37:37",
        "VOLUME,1.0,3014-321098,510100000000A9EF4X,20.1.26.52,新成彭路(洞子口路口至三环路口),2019-02-21 17:43:00,60,3,3,2B,3,0,2,1,0,0,22.0,0,3,105.0,2019-02-21 17:43:02",
        "VOLUME,1.0,3014-321099,510100000000A9EF4W,20.1.26.51,新成彭路(三环路口至洞子口路口),2019-02-21 17:43:00,60,3,3,2A,3,0,0,3,0,0,14.0,0,1,70.0,2019-02-21 17:43:02",
        "VOLUME,1.0,3014-321100,,20.3.134.55,成彭高速出成都14km+650m处,2019-02-21 17:44:00,60,6,-1,,5,0,0,5,0,0,5.0,0,2,12.0,2019-02-22 01:41:04",
        "VOLUME,1.0,3014-321079,510100000000A90080,20.3.134.16,入高速新繁收费站分流点2入高速,2019-02-21 17:44:00,60,2,2,2B,6,0,0,6,0,0,36.0,0,1,183.0,2019-02-22 01:42:16",
        "VOLUME,1.0,3014-321014,510100000000A32096,20.3.133.9,成彭高速公路（川高速S1）6公里919米（彭州至成都）,2019-02-21 17:44:00,60,3,5,2A,0,0,0,0,0,0,0.0,0,49,200.0,2019-02-21 17:38:37",
        "VOLUME,1.0,3014-321024,,20.3.133.39,成彭高速公路K5+200,2019-02-21 17:44:00,60,8,2,2A,0,0,0,0,0,0,0.0,0,0,200.0,2019-02-22 01:41:04",
        "VOLUME,1.0,3014-321056,510100000000A90057,20.3.133.17,入高速龙桥收费站分流点入高速,2019-02-21 17:44:00,60,2,1,,0,0,0,0,0,0,0.0,0,0,200.0,2019-02-22 01:42:16",
        "VOLUME,1.0,3014-321075,510100000000A90076,20.3.134.12,成彭高速公路（川高速S1）13公里228米（彭州至成都）,2019-02-21 17:44:00,60,4,1,2B,5,0,0,5,0,0,66.0,0,1,187.0,2019-02-22 01:42:16",
        "VOLUME,1.0,3014-321060,510100000000A90061,20.3.133.21,成彭高速公路（川高速S1）8公里205米（成都至彭州）,2019-02-21 17:44:00,60,3,6,2B,0,0,0,0,0,0,0.0,0,50,200.0,2019-02-22 01:42:16",
        "VOLUME,1.0,3014-321052,510100000000A90053,20.3.133.13,成彭高速公路(川高速S1)7公路590米(彭州至成都),2019-02-21 17:44:00,60,3,5,2A,18,0,0,18,0,0,0.0,0,11,0.0,2019-02-22 01:42:16",
        "VOLUME,1.0,3014-321090,510100000000A90093,20.3.134.35,成彭高速公路(川高速S1)17公里400米(彭州至成都),2019-02-21 17:44:00,60,2,1,2A,8,0,0,8,0,0,0.0,0,3,0.0,2019-02-22 01:42:16",
        "VOLUME,1.0,3014-321083,510100000000A90086,20.3.134.24,成彭高速公路（川高速S1）14公里760米（成都至彭州）,2019-02-21 17:44:00,60,3,4,2A,0,0,0,0,0,0,0.0,0,0,200.0,2019-02-22 01:42:16",
        "VOLUME,1.0,3014-321058,510100000000A90059,20.3.133.19,成彭高速公路(川高速S1)8公路120米(彭州至成都),2019-02-21 17:44:00,60,3,2,2B,7,0,0,7,0,0,0.0,0,1,0.0,2019-02-22 01:42:16"
		]
	
}`

		} else if cmd == vtd_type {
			resp_data = `[
    {
        "datasource": "0",
        "deviceid": "3050-321001",
        "devicepos": "成彭高速出成都14km+650m处",
        "laneno": "1",
        "eventtype": "E",
        "eventtime": "2019-03-05 08:20:09.364",
        "starttime": "2019-03-05 08:20:06.0",
        "endtime": "2019-03-05 08:20:07.0",
        "videostarttime": "2019-03-05 08:19:06.396",
        "videoendtime": "2019-03-05 08:22:06.396",
        "direction": null,
        "piccount": "2",
        "picx": "712",
        "picy": "1780",
        "isvalid": "1",
        "isabnormal": "0",
        "inserttime": "2019-03-05 08:20:09.0",
        "pic1": "http://20.5.11.42/SF_ITS_CP/SF_ITS_CP134/ALERT_PIC/20190305/08/3050-321050/E_3050-321050_20190305082009364_1.jpg",
        "pic2": "http://20.5.11.42/SF_ITS_CP/SF_ITS_CP134/ALERT_PIC/20190305/08/3050-321050/E_3050-321050_20190305082009364_2.jpg",
        "videopath": null,
        "id": "E_3050-321050_-1_20190305082009364"
    },
    {
        "datasource": "0",
        "deviceid": "3050-321007",
        "devicepos": "入高速龙桥收费站分流点入高速",
        "laneno": "4",
        "eventtype": "C",
        "eventtime": "2019-03-05 09:00:00.439",
        "starttime": "2019-03-05 08:59:43.0",
        "endtime": "2019-03-05 08:59:43.0",
        "videostarttime": "2019-03-05 08:58:43.425",
        "videoendtime": "2019-03-05 09:01:43.425",
        "direction": "2B",
        "piccount": "2",
        "picx": "1788",
        "picy": "864",
        "isvalid": "1",
        "isabnormal": "0",
        "inserttime": "2019-03-05 09:00:00.0",
        "pic1": "http://20.5.11.42/SF_ITS_CP/SF_ITS_CP133/ALERT_PIC/20190305/09/3014-321057/C_3014-321057_20190305090000439_1.jpg",
        "pic2": "http://20.5.11.42/SF_ITS_CP/SF_ITS_CP133/ALERT_PIC/20190305/09/3014-321057/C_3014-321057_20190305090000439_2.jpg",
        "videopath": null,
        "id": "C_3014-321057_4_20190305090000439"
    },
    {
        "datasource": "0",
        "deviceid": "3050-321004",
        "devicepos": "成彭高速出成都14km+650m处",
        "laneno": "-1",
        "eventtype": "E",
        "eventtime": "2019-03-05 08:33:42.579",
        "starttime": "2019-03-05 08:33:39.0",
        "endtime": "2019-03-05 08:33:40.0",
        "videostarttime": "2019-03-05 08:32:39.494",
        "videoendtime": "2019-03-05 08:35:39.494",
        "direction": null,
        "piccount": "2",
        "picx": "712",
        "picy": "1780",
        "isvalid": "1",
        "isabnormal": "0",
        "inserttime": "2019-03-05 08:33:42.0",
        "pic1": "http://20.5.11.42/SF_ITS_CP/SF_ITS_CP134/ALERT_PIC/20190305/08/3050-321050/E_3050-321050_20190305083342579_1.jpg",
        "pic2": "http://20.5.11.42/SF_ITS_CP/SF_ITS_CP134/ALERT_PIC/20190305/08/3050-321050/E_3050-321050_20190305083342579_2.jpg",
        "videopath": null,
        "id": "E_3050-321050_-1_20190305083342579"
    }
]`
		}

	}

	if cmd == vtd_type {
		resp_data = `[
    {
        "datasource": "0",
        "deviceid": "3050-321001",
        "devicepos": "成彭高速出成都14km+650m处",
        "laneno": "1",
        "eventtype": "E",
        "eventtime": "2019-03-05 08:20:09.364",
        "starttime": "2019-03-05 08:20:06.0",
        "endtime": "2019-03-05 08:20:07.0",
        "videostarttime": "2019-03-05 08:19:06.396",
        "videoendtime": "2019-03-05 08:22:06.396",
        "direction": null,
        "piccount": "2",
        "picx": "712",
        "picy": "1780",
        "isvalid": "1",
        "isabnormal": "0",
        "inserttime": "2019-03-05 08:20:09.0",
        "pic1": "http://dpic.tiankong.com/wf/mr/QJ6194637980.jpg",
        "pic2": "http://dpic.tiankong.com/xh/bl/QJ6313040190.jpg",
        "videopath": null,
        "id": "E_3050-321050_-1_20190305082009364"
    },
    {
        "datasource": "0",
        "deviceid": "3050-321007",
        "devicepos": "入高速龙桥收费站分流点入高速",
        "laneno": "4",
        "eventtype": "C",
        "eventtime": "2019-03-05 09:00:00.439",
        "starttime": "2019-03-05 08:59:43.0",
        "endtime": "2019-03-05 08:59:43.0",
        "videostarttime": "2019-03-05 08:58:43.425",
        "videoendtime": "2019-03-05 09:01:43.425",
        "direction": "2B",
        "piccount": "2",
        "picx": "1788",
        "picy": "864",
        "isvalid": "1",
        "isabnormal": "0",
        "inserttime": "2019-03-05 09:00:00.0",
        "pic1": "http://dpic.tiankong.com/w2/g2/QJ6398279548.jpg",
        "pic2": "http://dpic.tiankong.com/wf/mr/QJ6194637980.jpg",
        "videopath": null,
        "id": "C_3014-321057_4_20190305090000439"
    },
    {
        "datasource": "0",
        "deviceid": "3050-321004",
        "devicepos": "成彭高速出成都14km+650m处",
        "laneno": "-1",
        "eventtype": "E",
        "eventtime": "2019-03-05 08:33:42.579",
        "starttime": "2019-03-05 08:33:39.0",
        "endtime": "2019-03-05 08:33:40.0",
        "videostarttime": "2019-03-05 08:32:39.494",
        "videoendtime": "2019-03-05 08:35:39.494",
        "direction": null,
        "piccount": "2",
        "picx": "712",
        "picy": "1780",
        "isvalid": "1",
        "isabnormal": "0",
        "inserttime": "2019-03-05 08:33:42.0",
        "pic1": "http://dpic.tiankong.com/xh/bl/QJ6313040190.jpg",
        "pic2": "http://dpic.tiankong.com/i6/29/QJ6119124074.jpg",
        "videopath": null,
        "id": "E_3050-321050_-1_20190305083342579"
    }
]`
	}

	if cmd == mdi_type {
		resp_data = `{
    "status" : true, 
    "message" : "SUCCESS", 
    "data" : [
	{
		"SID" : "84446654",
		"TT" : "2018-12-18 15:22:10",
		"AA" : "339.0",
		"AB" : "11.0",
		"BD" : "70.0",
		"BP" : "9556.0",
		"BC" : "195.0",
		"APB" : "198",
		"HB1T" : "6292",
		"NL" : "228",
		"TEST" : ""
	},
	{
		"SID" : "C0001",
		"TT" : "2018-12-18 15:22:10",
		"AA" : "350.0",
		"AB" : "13.0",
		"BD" : "72.0",
		"BP" : "9550.0",
		"BC" : "196.0",
		"APB" : "190",
		"HB1T" : "6290",
		"NL" : "220",
		"TEST" : ""
	},
	{
		"SID" : "C0002",
		"TT" : "2018-12-18 15:22:10",
		"AA" : "333.0",
		"AB" : "10.0",
		"BD" : "74.0",
		"BP" : "9500.0",
		"BC" : "190.0",
		"APB" : "180",
		"HB1T" : "6200",
		"NL" : "200",
		"TEST" : ""
	}]
}`
	}

	logger.Info("get_state:%d start 1", cmd)

	go http_p.parse_http_data(cmd, resp_data, rep_cmd)
	logger.Info("get_state:%d end", cmd)
}

func (http_p *http_info) parse_http_data(cmd int, data, rep_cmd string) {
	var dev_str []byte
	switch cmd {
	case vld_type:
		dev_str, _ = http_p.parse_vld_data(data)
		//go platform_info.parse_vld_data(resp_data)
	case mdi_type:
		dev_str, _ = http_p.parse_mdi_data(data)
	case vtd_type:
		dev_str, _ = http_p.parse_vtd_data(data)
	default:

		break
	}

	//数据上报
	if dev_str != nil {
		go http_cms.Http_cms_p.Report_device_status(dev_str, rep_cmd)
	}
	logger.Info("get_state:%d len[%d] end", cmd, len(data))
}

func (http_p *http_info) http_request(url, token, body string) (string, error) {
	/*
		defer func() {
			if err := recover(); err != nil {
				log.Fatalf("Platform http_request error:[%v]", err)
				return
			}
		}()
	*/
	//resp, _ := http.Post(url, "application/x-www-form-urlencoded", strings.NewReader(body))
	//resp, err := http.PostForm(url, url.Values{"User-Access-Token": token})

	logger.Info("http url:%s start", url)
	//var request http.Request
	//c := make(chan struct{})
	//timer := time.AfterFunc(5*time.Second, func() {
	//	close(c)
	//})

	request, err := http.NewRequest(http.MethodPost, url, strings.NewReader(body))
	if err != nil {
		logger.Error("http NewRequest url:%s err:[%v]", url, err)
		return "", err
	}
	request.Header.Set("Content-Type", "application/json;charset=utf-8")
	request.Header.Set("User-Access-Token", token)
	//request.Cancel = c

	// client := &http.Client{
	// 	Timeout: 10 * time.Second,
	// }

	resp, err := http_p.client.Do(request)
	if err != nil {
		logger.Error("Client url:%s err:[%v]", url, err)
		return "", err
	}

	defer resp.Body.Close()
	rep_body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logger.Error("ReadAll url:%s err:[%v]", url, err)
		return "", err
	}

	// for {
	// 	timer.Reset(5 * time.Second)
	// 	// Try instead: timer.Reset(50 * time.Millisecond)
	// 	_, err = io.CopyN(ioutil.Discard, resp.Body, 256)
	// 	if err == io.EOF {
	// 		logger.Error("http_request io EOF")
	// 	} else if err != nil {
	// 		logger.Error("http_request err: %s", err.Error())
	// 	}
	// 	return "", err
	// }
	logger.Info("http url:%s end", url)
	//logger.Info("http url:%s resp:[%s]", url, string(rep_body))
	return string(rep_body), nil
}

func (http_p *http_info) parse_vld_data(data string) ([]byte, error) {
	defer func() {
		if err := recover(); err != nil {
			logger.Error("parse_vld_data error:[%v]", err)
			return
		}
	}()
	logger.Info("parse_vld_data start data:[%d]", len(data))
	resp_str := response_vld_str{}
	logger.Info("parse_vld_data start data1:[%d]", len(data))
	err := json.Unmarshal([]byte(data), &resp_str)
	if err != nil {
		logger.Error("parse_vld_data data invalid err:%v", err)
		return nil, err
	}
	logger.Info("parse_vld_data data len[%d]", len(resp_str.Data))

	rep_str := report_str{}
	rep_str.Gateway = config.Config_p.Gateway
	rep_str.Devicestatus = true
	rep_str.Internetstatus = true

	//data
	var vld_ids []string
	vld_data := vld_str{}
	for _, value := range resp_str.Data {
		v_str := strings.Split(value, ",")
		logger.Info("parse_vld_data v_str len[%d]", len(v_str))
		outer_id := v_str[2]
		Id, ret := http_p.find_valid_device(outer_id)
		if !ret {
			logger.Warn("vld outer_id[%s] invalid", outer_id)
			continue
		}
		logger.Info("parse_vld_data id[%s]", outer_id)

		vld_ids = append(vld_ids, Id)
		vld_data.Outerid = Id
		vld_data.Name = v_str[5]
		totalflow, _ := strconv.Atoi(string(v_str[11]))
		vld_data.Totalflow = totalflow
		vehiclebig, _ := strconv.Atoi(v_str[13])
		vld_data.Vehiclebig = vehiclebig
		vehiclemiddle, _ := strconv.Atoi(v_str[14])
		vld_data.Vehiclemiddle = vehiclemiddle
		vehicle1, _ := strconv.Atoi(v_str[12])
		vehicle2, _ := strconv.Atoi(v_str[15])
		vehicle3, _ := strconv.Atoi(v_str[16])
		vld_data.Vehiclesmall = vehicle1 + vehicle2 + vehicle3
		avgspeed, _ := strconv.ParseFloat(v_str[17], 32)
		vld_data.Avgspeed = float32(avgspeed)
		avgoccrate, _ := strconv.Atoi(v_str[19])
		vld_data.Avgoccrate = avgoccrate

		rep_str.Data = append(rep_str.Data, vld_data)
		logger.Info("parse_vld_data rep_str.Data[%d]", len(rep_str.Data))
	}

	//exception
	logger.Info("parse_vld_data exception start")
	if len(vld_ids) == 0 {
		logger.Warn("vld outer_id all invalid")
		return nil, errors.New("vld outer_id all invalid")
	}

	sort.Strings(vld_ids)
	for _, dev := range http_p.Devices {
		if dev.ResType == resType_vld {
			if http_p.find_exception_id(vld_ids, dev.Id) {
				logger.Warn("parse_vld_data exception id[%s]", dev.Id)
				rep_str.Exception = append(rep_str.Exception, dev.Id)
			}
		}
	}
	logger.Info("parse_vld_data exception end")
	str, err := json.Marshal(rep_str)
	if err != nil {
		logger.Error("parse_vld_data Marshal err: %s", err.Error())
		return nil, err
	}
	//logger.Info("parse_vld_data rep str: %s", str)
	logger.Info("parse_vld_data end rep[%d]: %s", len(str), str)

	return str, nil
}

func (http_p *http_info) parse_mdi_data(data string) ([]byte, error) {
	defer func() {
		if err := recover(); err != nil {
			logger.Error("parse_mdi_data error:[%v]", err)
			return
		}
	}()
	//logger.Info("parse_mdi_data start data:[%s]", data)
	rep_str := report_str{}
	rep_str.Gateway = config.Config_p.Gateway
	rep_str.Devicestatus = true
	rep_str.Internetstatus = true

	//data
	var mdi_ids []string
	resp_str := response_mdi_str{}
	err := json.Unmarshal([]byte(data), &resp_str)
	if err != nil {
		logger.Error("parse_mdi_data data invalid err:%v", err)
		return nil, err
	}
	logger.Info("parse_mdi_data len[%d]", len(resp_str.Data))
	for _, value := range resp_str.Data {
		sid := value.SID
		Id, ret := http_p.find_valid_device(sid)
		if !ret {
			logger.Warn("parse_mdi_data SID[%s] invalid", sid)
			continue
		}
		logger.Info("parse_mdi_data id[%s]", sid)
		var mdi_data mdi_str
		mdi_ids = append(mdi_ids, Id)
		mdi_data.Outerid = Id
		mdi_data.Coltime = value.TT
		Inswinddir, _ := strconv.ParseFloat(value.AA, 32)
		mdi_data.Inswinddir = float32(Inswinddir)
		Inswindspeed, _ := strconv.ParseFloat(value.AB, 32)
		mdi_data.Inswindspeed = float32(Inswindspeed)
		Humidity, _ := strconv.ParseFloat(value.BD, 32)
		mdi_data.Humidity = float32(Humidity)
		Airpressure, _ := strconv.ParseFloat(value.BP, 32)
		mdi_data.Airpressure = float32(Airpressure)
		Temperature, _ := strconv.ParseFloat(value.BC, 32)
		mdi_data.Temperature = float32(Temperature)
		Roadbedtemperature, _ := strconv.ParseFloat(value.APB, 32)
		mdi_data.Roadbedtemperature = float32(Roadbedtemperature)
		Visibility, _ := strconv.ParseFloat(value.HB1T, 32)
		mdi_data.Visibility = float32(Visibility)
		Roadsurfacetemperature, _ := strconv.ParseFloat(value.NL, 32)
		mdi_data.Roadsurfacetemperature = float32(Roadsurfacetemperature)

		rep_str.Data = append(rep_str.Data, mdi_data)
	}
	/*
		result := gjson.Get(data, "data.#")
		if result.Num <= 0 {
			logger.Error("parse_mdi_data data invalid:[%s]", data)
			return "", errors.New("parse_mdi_data data invalid!")
		}
		for _, value := range result.Array() {
			sid := value.Get("SID").String()
			Id, ret := http_cms.Http_cms_p.Find_valid_device(sid)
			if !ret {
				logger.Warn("mdi SID[%s] invalid", sid)
				continue
			}
			var mdi_data mdi_str
			mdi_ids = append(mdi_ids, Id)
			mdi_data.Outerid = Id
			mdi_data.Coltime = value.Get("TT").String()
			mdi_data.Inswinddir = float32(value.Get("AA").Float())
			mdi_data.Inswindspeed = float32(value.Get("AB").Float())
			mdi_data.Humidity = float32(value.Get("BD").Float())
			mdi_data.Airpressure = float32(value.Get("BP").Float())
			mdi_data.Temperature = float32(value.Get("BC").Float())
			mdi_data.Roadbedtemperature = float32(value.Get("APB").Float())
			mdi_data.Visibility = float32(value.Get("HB1T").Float())
			mdi_data.Roadsurfacetemperature = float32(value.Get("NL").Float())

			rep_str.Data = append(rep_str.Data, mdi_data)
		}*/

	//exception
	logger.Info("parse_mdi_data exception start")
	if len(mdi_ids) == 0 {
		logger.Warn("mdi outer_id all invalid")
		return nil, errors.New("mdi outer_id all invalid")
	}
	sort.Strings(mdi_ids)
	for _, dev := range http_p.Devices {
		if dev.ResType == resType_mdi {
			if http_p.find_exception_id(mdi_ids, dev.Id) {
				logger.Warn("parse_mdi_data exception id[%s]", dev.Id)
				rep_str.Exception = append(rep_str.Exception, dev.Id)
			}
		}
	}

	str, err := json.Marshal(rep_str)
	if err != nil {
		logger.Error("parse_mdi_data Marshal err: %s", err.Error())
		return nil, err
	}
	logger.Info("parse_mdi_data rep str[%d]: %s", len(str), str)

	return str, nil
}

func (http_p *http_info) parse_vtd_data(data string) ([]byte, error) {
	defer func() {
		if err := recover(); err != nil {
			logger.Error("parse_vtd_data error:[%v]", err)
			return
		}
	}()
	//logger.Info("parse_vtd_data start data[%s]", data)

	//data
	//response_vtd_str
	resp_str := []response_vtd_str{}
	err := json.Unmarshal([]byte(data), &resp_str)
	if err != nil {
		logger.Error("parse_vtd_data data invalid err:%v", err)
		return nil, err
	}

	if len(resp_str) <= 0 {
		logger.Error("parse_vtd_data data invalid is empty!")
		return nil, err
	}
	//logger.Info("parse_vtd_data resp_str len[%d]", len(resp_str))
	rep_str := report_str{}
	rep_str.Gateway = config.Config_p.Gateway
	rep_str.Devicestatus = true
	rep_str.Internetstatus = true

	for _, value := range resp_str {
		outer_id := value.Deviceid
		Id, ret := http_p.find_valid_device(outer_id)
		if !ret {
			logger.Warn("vtd outer_id[%s] invalid", outer_id)
			continue
		}
		logger.Info("parse_vtd_data id[%s]", outer_id)
		var vtd_data vtd_str
		vtd_data.Outerid = Id
		vtd_data.Event_type = "EVENT"
		vtd_data.Version = "1.0"
		Datasource, _ := strconv.Atoi(value.Datasource)
		vtd_data.Datasource = Datasource
		vtd_data.Devicelocation = value.Devicepos
		vtd_data.Eventcreatetime = value.Eventtime
		vtd_data.Eventstarttime = value.Starttime
		vtd_data.Eventendtime = value.Endtime
		Lanenumber, _ := strconv.Atoi(value.Laneno)
		vtd_data.Lanenumber = Lanenumber
		vtd_data.Eventtype = value.Eventtype
		vtd_data.Drection = value.Direction
		Eventlocationx, _ := strconv.Atoi(value.Picx)
		vtd_data.Eventlocationx = Eventlocationx
		Eventlocationy, _ := strconv.Atoi(value.Picy)
		vtd_data.Eventlocationy = Eventlocationy
		picNumber, _ := strconv.Atoi(value.Piccount)
		vtd_data.Picnumber = picNumber
		// for idx := 0; idx < picNumber; idx++ {
		// 	var name bytes.Buffer
		// 	fmt.Fprintf(&name, "pic%d", idx+1)
		// 	var pic_info pic
		// 	pic_info.Name = name.String()
		// 	//pic_info.Url = value.Get(name.String()).String() //?
		// 	vtd_data.Pics = append(vtd_data.Pics, pic_info)
		// }
		var pic_info pic
		pic_info.Name = "pic1"
		pic_info.Url = value.Pic1
		vtd_data.Pics = append(vtd_data.Pics, pic_info)
		pic_info.Name = "pic2"
		pic_info.Url = value.Pic2
		vtd_data.Pics = append(vtd_data.Pics, pic_info)

		vtd_data.Startvideotime = value.Videostarttime
		vtd_data.Endvideotime = value.Videoendtime
		vtd_data.Videourl = value.Videopath

		rep_str.Data = append(rep_str.Data, vtd_data)
	}
	/*
		var datas []interface{}
		_ = json.Unmarshal([]byte(data), &datas)

		for _, val := range datas {
			switch val_s := val.(type) {
			case map[interface{}]interface{}:
				//logger.Info("parse_vtd_data val_s:%s", val_s)
				for _, val_st := range val_s {
					switch val_str := val_st.(type) {
					case map[string]interface{}:
						for v, val_str_ := range val_str {
							logger.Info("parse_vtd_data %s val_str_: %s", v, val_str_.(string))
						}
					default:
						logger.Error("parse_vtd_data val_str invalid %v", val_str)
					}
				}
				//outer_id := gjson.Get(val_s, "deviceid").String()
				//logger.Info("parse_vtd_data val_s deviceid:%s", outer_id)
			default:
				logger.Error("parse_vtd_data val_s invalid %v", val_s)
			}
		}*/
	/*
		result := gjson.Get(data, "#")
		logger.Info("parse_vtd_data result:%s", result.String())
		if result.Num <= 0 {
			logger.Error("parse_vtd_data data invalid:[%s]", data)
			return "", errors.New("parse_vtd_data data invalid!")
		}
		for _, value := range result.Array() {

			outer_id := value.Get("deviceid").String()
			Id, ret := http_cms.Http_cms_p.Find_valid_device(outer_id)
			if !ret {
				logger.Warn("vtd outer_id[%s] invalid", outer_id)
				continue
			}
			var vtd_data vtd_str
			vtd_data.Outerid = Id
			vtd_data.Event_type = "EVENT"
			vtd_data.Version = "1.0"
			vtd_data.Datasource = int(value.Get("datasource").Int())
			vtd_data.Devicelocation = value.Get("devicepos").String()
			vtd_data.Eventcreatetime = value.Get("eventtime").String()
			vtd_data.Eventstarttime = value.Get("starttime").String()
			vtd_data.Eventendtime = value.Get("endtime").String()
			vtd_data.Lanenumber = int(value.Get("laneno").Int())
			vtd_data.Eventtype = value.Get("eventtype").String()
			vtd_data.Drection = value.Get("direction").String()
			vtd_data.Lanenumber = int(value.Get("devicepos").Int())
			vtd_data.Eventlocationx = int(value.Get("picx").Int())
			vtd_data.Eventlocationy = int(value.Get("picy").Int())
			picNumber := int(value.Get("piccount").Int())
			vtd_data.Picnumber = picNumber
			for idx := 0; idx < picNumber; idx++ {
				var name bytes.Buffer
				fmt.Fprintf(&name, "pic%d", idx+1)
				var pic_info pic
				pic_info.Name = name.String()
				pic_info.Url = value.Get(name.String()).String()
				vtd_data.Pics = append(vtd_data.Pics, pic_info)
			}

			rep_str.Data = append(rep_str.Data, vtd_data)
		}*/

	str, err := json.Marshal(rep_str)
	if err != nil {
		logger.Error("parse_vtd_data Marshal err: %s", err.Error())
		return nil, err
	}
	logger.Info("parse_vtd_data rep str[%d]: %s", len(str), str)

	return str, nil
}

func (http_p *http_info) find_valid_device(outer_id string) (string, bool) {

	for _, dev := range http_p.Devices {
		if dev.OuterId == outer_id {
			return dev.Id, true
		}
	}
	return "", false
}

func (http_p *http_info) find_exception_id(ids []string, id string) bool {
	for _, value := range ids {
		if value == id {
			return false
		}
	}
	return true
}
