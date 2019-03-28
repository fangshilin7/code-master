//通过etcd管理config
package etcd

import (
	"context"
	"fmt"
	"git.scsv.online/go/base/logger"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"time"
)

//配置文件
type baseConfig struct {
	Etcd string `yaml:"etcd"`
	Type string `yaml:"type"`
	SN   string `yaml:"sn"`
}

var ec *Etcd
var cfgPath string

//加载配置
func LoadConfig(path string, out interface{}, autoupdate bool) (ecd *Etcd, err error) {
	cfgPath = path

	content, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}

	cfg := &baseConfig{}
	err = yaml.Unmarshal(content, &cfg)
	if err != nil {
		return
	}

	err = yaml.Unmarshal(content, out)
	if err != nil {
		return
	}

	//未配置etcd
	if cfg.Etcd == "" || cfg.Type == "" || cfg.SN == "" {
		err = yaml.Unmarshal(content, out)
		return
	}

	//创建etcd
	ec, err = NewEtcd(cfg.Etcd)
	if err != nil {
		return
	}

	ecd = ec
	key := fmt.Sprintf("/config/%s/%s", cfg.Type, cfg.SN)
	err = ec.SetKey(key, string(content), INFINITE)
	if err != nil {
		return
	}

	if autoupdate {
		go autoUpdate(key)
	}

	return
}

//自动根据etcd更新配置
func autoUpdate(key string) {
	w := ec.Watcher(key)
	for {
		//logger.Trace("etcd watch %s", key)
		r, err := w.Next(context.Background())
		if err != nil {
			//logger.Trace(err.Error())
			<-time.After(time.Second * 10)
			continue
		}

		if r.Action == "set" {
			logger.Trace("Update Config, %s => \n%s", key, r.Node.Value)
			err = ioutil.WriteFile(cfgPath, []byte(r.Node.Value), 0666)
			if err != nil {
				logger.Error(err.Error())
			}
		} else {
			logger.Trace("watcher %s, %+v", r.Action, r.Node)
		}
	}
}
