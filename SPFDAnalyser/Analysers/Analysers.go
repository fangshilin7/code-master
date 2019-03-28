//  SPFDAnalyser project Analyser_open.go

package main

import (
	"Analyser"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"

	//"time"

	"gopkg.in/yaml.v2"
)

type config struct {
	Device_id_start int
	Device_count    int
	Service_addr    string
}

func main() {

	var config_info config
	path, _ := os.Getwd()
	data, err := ioutil.ReadFile(path + "/config.yml")
	if err != nil {
		fmt.Println("load config err", err)
		return
	}

	//fmt.Println(string(data))
	err = yaml.Unmarshal(data, &config_info)
	if err != nil {
		fmt.Println("yaml Unmarshal err", err)
		return
	}
	fmt.Println("addr:", config_info.Service_addr)

	pkt_len, _ := Analyser.Simulation_data_load()

	runtime.GOMAXPROCS(runtime.NumCPU())
	for i := config_info.Device_id_start; i < config_info.Device_count; i++ {
		str_id := bulid_id(i)
		fmt.Println(str_id)
		analyser := Analyser.Analyser{}
		err := analyser.Start(config_info.Service_addr, str_id, pkt_len)
		if err != nil {
			fmt.Println("analyser Start err", err)
			continue
		}
		//go analyser.Run()
		//time.Sleep(time.Duration(50) * time.Millisecond)
	}

	fmt.Println("Hello World!", pkt_len)

	select {}
}

func bulid_id(id int) string {
	var str_id string
	if 0 <= id && id < 10 {
		str_id = fmt.Sprintf("024355483000%d", id)
	} else if 10 <= id && id < 100 {
		str_id = fmt.Sprintf("02435548300%d", id)
	} else if 100 <= id && id < 1000 {
		str_id = fmt.Sprintf("0243554830%d", id)
	} else if 1000 <= id && id < 10000 {
		str_id = fmt.Sprintf("024355483%d", id)
	}
	return str_id
}
