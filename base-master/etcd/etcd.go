package etcd

import (
	"context"
	"git.scsv.online/go/base/logger"
	"go.etcd.io/etcd/client"
	"time"
)

type Etcd struct {
	url  string
	kapi client.KeysAPI
}

const INFINITE = 0xffffffff

//创建全局etcd
func NewEtcd(url string) (*Etcd, error) {
	cfg := client.Config{
		Endpoints:               []string{url},
		Transport:               client.DefaultTransport,
		HeaderTimeoutPerRequest: time.Second * 2,
	}
	c, err := client.New(cfg)
	if err != nil {
		logger.Error(err.Error())
		return nil, err
	}

	etcd := &Etcd{
		url:  url,
		kapi: client.NewKeysAPI(c),
	}

	return etcd, nil
}

func (etcd *Etcd) SetKey(path string, val string, ttl int) (err error) {
	logger.Trace("ETCD SetKey, %s => %s, ttl = %d", path, val, ttl)
	if ttl == INFINITE {
		_, err = etcd.kapi.Set(context.Background(), path, val, nil)
	} else {
		_, err = etcd.kapi.Set(context.Background(), path, val,
			&client.SetOptions{TTL: time.Second * time.Duration(ttl)})
	}

	if err != nil {
		logger.Error(err.Error())
	}
	return
}

func (etcd *Etcd) GetValue(path string) (val string, err error) {
	var resp *client.Response
	resp, err = etcd.kapi.Get(context.Background(), path, nil)
	if err != nil {
		logger.Trace(err.Error())
		return
	}
	if resp.Node != nil {
		val = resp.Node.Value
		logger.Trace("ETCD GetValue, %s => %s", path, val)
	}

	return
}

func (etcd *Etcd) GetKeys(path string) (keys []string, err error) {
	var resp *client.Response
	resp, err = etcd.kapi.Get(context.Background(), path, nil)
	if err != nil {
		logger.Trace(err.Error())
		return
	}
	if resp.Node == nil || resp.Node.Nodes == nil {
		return
	}

	for _, v := range resp.Node.Nodes {
		keys = append(keys, v.Key)
	}

	logger.Trace("ETCD GetKeys, %s => %v", path, keys)
	return
}

func (etcd *Etcd) GetKeyValues(path string) (keys []string, values []string, err error) {
	var resp *client.Response
	resp, err = etcd.kapi.Get(context.Background(), path, nil)
	if err != nil {
		logger.Trace(err.Error())
		return
	}
	if resp.Node == nil || resp.Node.Nodes == nil {
		return
	}

	for _, v := range resp.Node.Nodes {
		keys = append(keys, v.Key)
		values = append(values, v.Value)
	}

	logger.Trace("ETCD GetKeys, %s => %v, %v", path, keys, values)
	return
}

func (etcd *Etcd) DeleteKey(path string, dir bool) (err error) {
	_, err = etcd.kapi.Delete(context.Background(), path, &client.DeleteOptions{Dir: dir, Recursive: true})
	if err != nil {
		logger.Error(err.Error())
		return
	}

	logger.Trace("ETCD DeleteKey, %s ", path)
	return
}

func (etcd *Etcd) Watcher(key string) client.Watcher {
	return etcd.kapi.Watcher(key, &client.WatcherOptions{Recursive: true})
}
