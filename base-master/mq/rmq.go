package mq

import (
	"encoding/json"
	"git.scsv.online/go/base/logger"
	"github.com/streadway/amqp"
)

type RMQ struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	// 服务器地址
	url      string
	kind     string
	exchange string
}

func NewRMQ(url, exchange, kind string) (*RMQ, error) {
	obj := &RMQ{
		url:      url,
		kind:     kind,
		exchange: exchange,
	}
	
	// 连接服务器
	err := obj.Connect()
	if err != nil {
		return nil, err
	}

	return obj, nil
}

// 连接服务器
func (obj *RMQ) Connect() error {
	var err error
	conn, err := amqp.Dial(obj.url)

	// 连接失败
	if err != nil {
		return err
	}

	// 打开通道
	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return err
	}

	// 声明exchange
	err = channel.ExchangeDeclare(obj.exchange, obj.kind, true, false, false, true, nil)
	if err != nil {
		obj.DisConnect()
		return err
	}

	obj.conn = conn
	obj.channel = channel
	logger.Info("connect to rabbitmq：%s", obj.url)

	return nil
}

// 断开连接
func (obj *RMQ) DisConnect() {
	logger.Info("close rabbitmq connection")

	// 关闭通道
	if obj.channel != nil {
		obj.channel.Close()
	}

	// 关闭连接
	if obj.conn != nil {
		obj.conn.Close()
	}
}

// 声明队列
func (obj *RMQ) Queue(name string, key string) error {
	_, err := obj.channel.QueueDeclare(name, true, true, false, true, nil)
	if err != nil {
		return err
	}

	// 绑定队列
	return obj.channel.QueueBind(name, key, obj.exchange, false, nil)
}

// 发布消息
func (obj *RMQ) Send(key string, data interface{}) {
	if obj == nil {
		return
	}

	if len(key) > 255 {
		logger.Error("rmq message length error, %v", data)
		return
	}

	logger.Trace("rmq message: %s->%#v", key, data)

	body, _ := json.Marshal(data)
	//body := data.([]byte)

	err := obj.channel.Publish(obj.exchange, key, true, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
		Expiration:  "120000",
	})

	if err != nil {
		logger.Error("send mq message error: %s", err.Error())
	}
}

// 订阅消息
func (obj *RMQ) Describe(key string, f func(key string, msg []byte)) {
	// 获取消费通道
	obj.channel.Qos(1, 0, true) // 确保rabbitmq会一个一个发消息
	msgs, err := obj.channel.Consume(key, "", true, false, false, false, nil)
	if err != nil {
		logger.Error("describe %s error(%s)", key, err.Error())
		return
	}

	if f != nil {
		for msg := range msgs {
			f(msg.RoutingKey, msg.Body)
		}
	}
}
