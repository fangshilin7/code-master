package config

import (
	//"fmt"
	"io/ioutil"
	"os"

	"git.scsv.online/go/logger"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Type     string
	Gateway  string
	Ip       string
	Protocol string
	Http     int
	Cms      string
	Ops      string
	Platform struct {
		Vld device_type
		Mdi device_type
		Vtd device_type
	}
}

type device_type struct {
	Url   string
	Token string
}

func (config *Config) Load() error {
	path, _ := os.Getwd()
	//fmt.Println(string(path))

	data, err := ioutil.ReadFile(path + "/config.yml")
	if err != nil {
		return err
	}

	//logger.Info("config: [%s]", string(data))
	err = yaml.Unmarshal(data, config)
	if err != nil {
		return err
	}
	logger.Info("config:[ \n\t%s\n\t%s\n\t%s\n\t%s\n\t%s\n\t%s\n\t%s]\n", (*config).Gateway, (*config).Ip,
		(*config).Cms, (*config).Ops, (*config).Platform.Vld.Url, (*config).Platform.Mdi.Url, (*config).Platform.Vtd.Url)
	return nil
}

var Config_p *Config
