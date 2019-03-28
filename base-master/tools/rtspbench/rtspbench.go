package main

import (
	"fmt"
	"git.scsv.online/go/base/config"
	"git.scsv.online/go/base/logger"
	"git.scsv.online/go/base/rtsp"
	"sync"
	"time"
)

type Source struct {
	Url     string `yaml:"url"`
	Num     int    `yaml:"num"`
	inbytes int
	sess    []*rtsp.ClientSession
}

type Config struct {
	Sources []*Source `yaml:"sources"`
	Time    int       `yaml:"time"`
	Max     int       `yaml:"max"`
}

var cfg Config
var count int

func (r *Source) OnPacket(packet *rtsp.RtspPacket) (err error) {
	if packet.Stream != nil {
		r.inbytes += len(packet.Stream)
		//fmt.Printf("%s %d\n", r.name, r.inbytes)
	}
	packet.Reset(nil)
	return nil
}
func (r *Source) OnError(packet *rtsp.RtspPacket, err error) {

}
func (r *Source) PacketPool() *sync.Pool {
	return nil
}

func closeSource() {
	for _, source := range cfg.Sources {
		if source.sess == nil {
			continue
		}
		for i, s := range source.sess {
			if s != nil {
				s.Close()
				source.sess[i] = nil
			}
		}
	}
}

func request() {
	for _, source := range cfg.Sources {
		if source.sess == nil {
			source.sess = make([]*rtsp.ClientSession, 5000)
		}

		for j := 0; j < source.Num; j++ {
			ch := make(chan int)
			sess, err := rtsp.Connect(source.Url, source, ch)
			if err != nil {
				fmt.Println(err)
				continue
			}

			<-ch
			source.sess[j] = sess
		}
	}
}

func main() {
	logger.LogToFile = false
	err := config.Load("rtspbench.yml", &cfg)
	if err != nil {
		logger.Error(err.Error())
		return
	}

	fmt.Println(cfg.Sources)

	go request()
	go func() {
		for {
			time.Sleep(time.Second * 5)
			var total int
			for i, s := range cfg.Sources {
				mbps := (s.inbytes * 8) / (1024 * 1024 * 5)
				logger.Info("Source[%d]: %d mbps", i, mbps)
				total += s.inbytes
				s.inbytes = 0
			}
			logger.Info("--- Total: %d mbps --- \r\n", (total*8)/(1024*1024*5))
		}
	}()

	if cfg.Time > 0 {
		go func() {
			for {
				<-time.After(time.Second * time.Duration(cfg.Time))
				logger.Info("Cycle Request")
				closeSource()
				<-time.After(time.Second * 1)
				request()
				count++
			}
		}()
	}

	if cfg.Max > 0 {
		for {
			if count > cfg.Max {
				break
			}
			time.Sleep(time.Second * 5)
		}
	} else {
		c := make(chan int)
		<-c
	}
}
