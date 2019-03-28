package cache

import (
	"encoding/json"
	"errors"
	"fmt"
	"git.scsv.online/go/base/logger"
	"github.com/gomodule/redigo/redis"
	"strings"
	"time"
)

type RedisCache struct {
	network string
	addr    string
	db      uint

	// 默认过期时间
	expire int

	pool *redis.Pool
}

func NewRedis(network string, addr string, db uint) (*RedisCache, error) {
	obj := &RedisCache{
		network: network,
		addr:    addr,
		db:      db,
		expire:  86400,
		pool: &redis.Pool{
			MaxIdle:     2,
			MaxActive:   5,
			IdleTimeout: 240 * time.Second,
			Dial: func() (redis.Conn, error) {
				c, err := redis.Dial(network, addr)
				if err != nil {
					return nil, err
				}

				// 选择db
				c.Do("SELECT", db)
				return c, nil
			},
		},
	}

	// 连接服务器
	err := obj.Connect()
	if err != nil {
		return nil, err
	}
	
	return obj, nil
}

/***
** purpose：
**   连接服务器
**
** params：
**   expireCb: 过期事件回调
**
** return:
**   error: 错误码
***/
// 连接服务器
func (obj *RedisCache) Connect() error {
	c := obj.pool.Get()
	if c.Err() != nil {
		return c.Err()
	}

	defer c.Close()
	logger.Info("connected redis server(%v)", obj.addr)

	return nil
}

// 订阅过期事件
func (obj *RedisCache) ExpireEvent(cb func(key string)) error {
	c := obj.pool.Get()
	if c.Err() != nil {
		return c.Err()
	}

	// 订阅过期事件
	psc := redis.PubSubConn{c}
	err := psc.PSubscribe(fmt.Sprintf("__keyevent@%d__:expired", obj.db))
	if err != nil {
		return err
	}

	for {
		switch v := psc.Receive().(type) {
		case redis.Message:
			cb(string(v.Data))
		case error:
			logger.Error("%v", v)
			return nil
		}
	}

	return nil
}

// 断开连接
func (obj *RedisCache) DisConnect() {
	logger.Info("close redis connection")
	obj.pool.Close()
}

// 设置缓存
func (obj *RedisCache) SetCache(key string, value interface{}) {
	obj.SetCacheEx(key, value, obj.expire)
}

// 执行命令
func (obj *RedisCache) do(cmd string, params ...interface{}) (interface{}, error) {
	// 记录开始时间
	start := time.Now().UnixNano()

	c := obj.pool.Get()
	if c.Err() != nil {
		return nil, c.Err()
	}

	defer c.Close()

	// key
	key := params[0]

	// 执行命令
	v, err := c.Do(cmd, params...)

	if err != nil {
		logger.Error("%s cache %s, error: %s", cmd, key, err.Error())
	} else if v == nil {
		err = errors.New(fmt.Sprintf("%v not exist", key))
	} else {
		logger.Trace("%s cache %v use %v ms", cmd, key, (time.Now().UnixNano()-start)/1e6)
	}

	return v, err
}

// 设置缓存
func (obj *RedisCache) SetCacheEx(key string, value interface{}, expire int) {
	// 数据格式化
	switch value.(type) {
	case string:
		break
	default:
		value, _ = json.Marshal(value)
		break
	}

	//data, _ := json.Marshal(value)
	obj.do("SET", key, value, "EX", expire)
}

// 查询缓存
func (obj *RedisCache) GetCache(key string) (interface{}, error) {
	return obj.do("GET", key)
}

// 查询索引
func (obj *RedisCache) GetKeys(key string) (interface{}, error) {
	return obj.do("KEYS", key)
}

// 刷新缓存过期时间
func (obj *RedisCache) SetExpire(key string, value ...int) {
	// 默认过期时间，24小时
	expire := obj.expire

	if len(value) > 0 {
		expire = value[0]
	}

	obj.do("EXPIRE", key, expire)
}

// 删除缓存
func (obj *RedisCache) DelCache(key string) (interface{}, error) {
	return obj.do("DEL", key)
}

/***
** purpose：
**   批量删除
**
** params：
**   key: 关键字
**   exclude: 排除项
**
** return:
**   error: 错误码
***/
func (obj *RedisCache) BatchDelMatchCache(key string, exclude []string) error {
	// 记录开始时间
	start := time.Now().UnixNano()

	// 获取可用连接
	c := obj.pool.Get()
	if c.Err() != nil {
		return c.Err()
	}

	defer c.Close()

	// 执行命令
	v, err := c.Do("KEYS", key)
	if err != nil {
		logger.Error(err.Error())
		return nil
	}

	keys := v.([]interface{})
	// 删除项
	var delKeys []interface{}
	// 无过滤项
	if len(exclude) == 0 {
		delKeys = keys
	} else {
		for _, key := range keys {
			k := string(key.([]byte))

			for _, ex := range exclude {
				// 非保留项
				if !strings.HasPrefix(k, ex) {
					delKeys = append(delKeys, k)
				}
			}
		}
	}

	// 删除
	if len(delKeys) > 0 {
		c.Do("DEL", delKeys...)
	}

	logger.Debug("del cache(%v) use %vms", delKeys, (time.Now().UnixNano()-start)/1e6)

	return nil
}

func (obj *RedisCache) BatchDelCache(keys []interface{}) error {
	// 删除
	if len(keys) > 0 {
		obj.do("DEL", keys...)
	}

	return nil
}
