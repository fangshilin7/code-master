package main

import (
	"git.scsv.online/go/base/logger"
	"git.scsv.online/go/base/mq"
)

var srvMQ *mq.RMQ

// 接收mq消息
func receive(key string, data []byte) {
	logger.Error("%s %s", key, data)

	//if key == "gw.track" {
	//	var track request.Track
	//
	//	// 数据格式化
	//	err := json.Unmarshal(data, &track)
	//	if err != nil {
	//		logger.Error("track msg(%s) error(%s)", data, err.Error())
	//		return
	//	}
	//
	//	//logger.Error("%s", track.Time.Format(dtype.Time_Layout))
	//} else {
	//	logger.Error("%s", data)
	//}

	//logger.Error("%s->%s", key, string(data))
}

func main() {
	// 消息生产者
	//srvMQ = mq.NewRMQ("amqp://guest:guest@192.168.1.186:10672?heartbeat=60", "2k1v.topic", "topic")
	//err := srvMQ.Connect()
	//if err != nil {
	//	logger.Error(err.Error())
	//	return
	//}
	//defer srvMQ.DisConnect()

	//// 消息分发
	//go func() {
	//	for {
	//		<- time.After(time.Second * 1)
	//		srvMQ.Send("t*", time.Now())
	//	}
	//}()

	// 消息消费者
	cli := mq.NewRMQ("amqp://guest:guest@192.168.1.186:10672?heartbeat=60", "2k1v.topic", "topic")
	err := cli.Connect()
	if err != nil {
		logger.Error(err.Error())
		return
	}
	defer cli.DisConnect()

	// 队列
	queue := "client"
	//err = cli.Queue(queue, "gw.#")
	err = cli.Queue(queue, "state.#")
	if err != nil {
		logger.Error(err.Error())
	}

	go cli.Describe(queue, receive)

	exit := make(chan bool)
	<-exit
}
