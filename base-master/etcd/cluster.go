package etcd

import (
	"fmt"
	"git.scsv.online/go/base/logger"
	"time"
)

type Cluster struct {
	flag     string
	ip       string
	port     int
	value    string
	closed   bool
	etcd     *Etcd
	nodepath string
	keypath  string
}

const CLUSTER_HEARTBEAT_TIME = 60

func NewCluster(etcd string, flag string, ip string, port int, value string) (cluster *Cluster, err error) {
	cluster = &Cluster{
		flag:     flag,
		ip:       ip,
		port:     port,
		value:    value,
		nodepath: fmt.Sprintf("/cluster/%s/nodes/%s:%d", flag, ip, port),
		keypath:  fmt.Sprintf("/cluster/%s/keys/", flag),
	}
	cluster.etcd, err = NewEtcd(etcd)
	if err != nil {
		logger.Error(err.Error())
		return
	}
	err = cluster.etcd.SetKey(cluster.nodepath, cluster.value, CLUSTER_HEARTBEAT_TIME)
	if err != nil {
		logger.Error(err.Error())
		return
	}

	go cluster.heartbeat()

	return
}

func (cluster *Cluster) Close() {
	cluster.etcd.DeleteKey(cluster.nodepath, false)
	cluster.closed = true
}

func (cluster *Cluster) SetKey(key string, value string, ttl int) (err error) {
	err = cluster.etcd.SetKey(cluster.keypath+key, value, ttl)
	return
}

func (cluster *Cluster) DeleteKey(key string) (err error) {
	err = cluster.etcd.DeleteKey(cluster.keypath+key, false)
	return
}

func (cluster *Cluster) GetValue(key string) (val string, err error) {
	val, err = cluster.etcd.GetValue(cluster.keypath + key)
	return
}

//获取所有节点
func (cluster *Cluster) GetKeyValues() (keys []string, values []string) {
	keys, values, _ = cluster.etcd.GetKeyValues("/cluster/ms/nodes")
	return
}

func (cluster *Cluster) heartbeat() {
	etcd := cluster.etcd
	val := cluster.value
	if val == "" {
		val = "OK"
	}
	for {
		<-time.After(time.Second * (CLUSTER_HEARTBEAT_TIME - 5))

		///cluster/ms/nodes/192.168.1.100:13002
		if cluster.closed {
			break
		}

		err := etcd.SetKey(cluster.nodepath, val, CLUSTER_HEARTBEAT_TIME)

		if err != nil {
			logger.Error(err.Error())
		}
	}
}
