package db

import (
	"errors"
	"git.scsv.online/go/base/logger"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"reflect"
	"time"
)

// 查询参数
type FindParam struct {
	// 查询条件
	Query interface{}
	// 查询选项
	Option interface{}
	// 排序
	Sort string
	// 翻页
	Page int
	// 查询条数
	Limit int
	// 数量统计
	Count bool
}

type MgoDB struct {
	url string
	db  string
	usr string
	pwd string

	client  *mgo.Database
	session *mgo.Session
}

func NewMgoDB(url string, database string) (*MgoDB, error) {
	obj := &MgoDB{
		url: url,
		db:  database,
	}
	
	// 连接服务器
	err := obj.Connect()
	if err != nil {
		return nil, err
	}
	
	return obj, nil
}

// 配置用户名密码
func (obj *MgoDB) Auth(usr, pwd string) error {
	obj.usr = usr
	obj.pwd = pwd
	return obj.client.Login(usr, pwd)
}

// 连接服务器
func (obj *MgoDB) Connect() error {
	session, err := mgo.Dial(obj.url)

	// 连接失败
	if err != nil {
		return err
	}

	// 连接数据库
	client := session.DB(obj.db)
	obj.session = session
	obj.client = client
	logger.Info("connected mongodb(%v) database(%v)", obj.url, obj.db)

	return nil
}

// 连接状态
func (obj *MgoDB) ConnectState() bool {
	return obj.session == nil
}

// 断开连接
func (obj *MgoDB) DisConnect() {
	logger.Info("close mongodb connection")

	if obj.session != nil {
		obj.session.Close()
	}
}

// 返回数据库实例
func (obj *MgoDB) GetDB() *mgo.Database {
	return obj.client
}

// 查询函数
func (obj *MgoDB) Find(collection string, param *FindParam, out interface{}) (int, error) {
	// 记录请求开始时间
	start := time.Now().UnixNano()
	// 获取集合
	c := obj.client.C(collection)
	var err error
	var count int

	// 集合不存在
	if c == nil {
		err = errors.New(collection + " is not exist.")
		return count, err
	}

	var query *mgo.Query

	// 自定义条件查询
	query = c.Find(param.Query).Select(param.Option)
	// 数量统计
	if param.Count {
		count, _ = query.Count()
	}

	// 排序
	if param.Sort != "" {
		query = query.Sort(param.Sort)
	}

	// 查询条数
	if param.Limit != 0 {
		// 翻页
		if param.Page > 0 {
			query = query.Skip(param.Limit*(param.Page - 1))
		}

		query = query.Limit(param.Limit)
	}

	// 输出参数类型
	t := reflect.TypeOf(out)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// 查询选项
	if param.Option != nil {
		query = query.Select(param.Option)
	}

	if t.Kind() == reflect.Slice {
		err = query.All(out)
	} else {
		err = query.One(out)
	}

	logger.Debug("MONGO %v.find, param: %#v, use %#vms",
		collection, param, (time.Now().UnixNano()-start)/1e6)

	return count, err
}

// 根据Id查询
func (obj *MgoDB) GetById(out interface{}, collection string, id ObjectId, option ...interface{}) error {
	param := &FindParam{Query: bson.M{"_id": id, }}

	// 查询选项
	if len(option) > 0 {
		param.Option = option[0]
	}

	_, err := obj.Find(collection, param, out)
	return err
}

// 插入数据
//   collection 集合名
//   doc 数据
func (obj *MgoDB) Insert(collection string, doc ...interface{}) (err error) {
	// 记录请求开始时间
	start := time.Now().UnixNano()
	// 获取集合
	c := obj.client.C(collection)

	// 集合不存在
	if c == nil {
		err = errors.New(collection + " is not exist.")
		return err
	}

	err = c.Insert(doc...)
	logger.Debug("MONGO %v.insert {\"document\": %#v} %vms", collection, doc, (time.Now().UnixNano()-start)/1e6)

	return err
}

// 更新-插入
//   collection 集合名
//   doc 数据
func (obj *MgoDB) Upsert(collection string, cond, doc interface{}) (err error) {
	// 记录请求开始时间
	start := time.Now().UnixNano()
	// 获取集合
	c := obj.client.C(collection)

	// 集合不存在
	if c == nil {
		err = errors.New(collection + " is not exist.")
		return err
	}

	_, err = c.Upsert(cond, doc)
	logger.Debug("MONGO %v.upsert {\"document\": %#v} %vms", collection, doc, (time.Now().UnixNano()-start)/1000000)

	return err
}

// 更新数据
//   collection 集合名
//   query 查询条件
//   doc 更新内容
//   unset 删除字段
//   multi 更新多条记录
func (obj *MgoDB) Update(collection string, query interface{}, doc interface{}, unset, multi bool) (err error) {
	var op string

	// 更新字段
	op = "$set"
	if unset {
		op = "$unset"
	}

	return obj.Update2(collection, query, doc, op, multi)
}

func (obj *MgoDB) Update2(collection string, query interface{}, doc interface{}, op string, multi bool) error {
	var err error

	// 记录请求开始时间
	start := time.Now().UnixNano()
	// 获取集合
	c := obj.client.C(collection)

	// 集合不存在
	if c == nil {
		err = errors.New(collection + " is not exist.")
		return err
	}

	// 更新操作
	data := bson.M{op: doc}

	// 更新多条记录
	if multi {
		_, err = c.UpdateAll(query, data)
	} else {
		err = c.Update(query, data)
	}

	logger.Debug("MONGO %v.update {\"condition\": %#v \"document\": %#v} %vms", collection, query, doc, (time.Now().UnixNano()-start)/1000000)

	return err
}

// 根据Id更新数据
//   collection 集合名
//   id document _id
//   doc 更新内容
func (obj *MgoDB) UpdateById(collection string, id interface{}, doc interface{}, unset bool) (err error) {
	return obj.Update(collection, bson.M{"_id": id}, doc, unset, false)
}

// 删除数据
//   collection 集合名
//   query 查询条件
func (obj *MgoDB) Remove(collection string, query interface{}) error {
	// 记录请求开始时间
	start := time.Now().UnixNano()
	// 获取集合
	c := obj.client.C(collection)

	// 集合不存在
	if c == nil {
		return errors.New(collection + " is not exist.")
	}

	_, err := c.RemoveAll(query)
	logger.Debug("MONGO %v.remove {\"condition\": %#v} %vms", collection, query, (time.Now().UnixNano()-start)/1000000)

	return err
}

// 删除数据
//   collection 集合名
//   id document _id
func (obj *MgoDB) RemoveById(collection string, id interface{}) error {
	return obj.Remove(collection, bson.M{"_id": id})
}

// 统计
//   collection 集合名
//   query 查询条件
func (obj *MgoDB) Count(collection string, cond interface{}) (int, error) {
	// 记录请求开始时间
	start := time.Now().UnixNano()
	// 获取集合
	c := obj.client.C(collection)

	// 集合不存在
	if c == nil {
		return 0, errors.New(collection + " is not exist.")
	}

	count, err := c.Find(cond).Count()
	logger.Debug("MONGO %v.count {\"conditions\": %#v} %vms", collection, cond, (time.Now().UnixNano()-start)/1e6)

	return count, err
}
