package db

import (
	"encoding/json"
	"fmt"
	"git.scsv.online/go/base/logger"
	"github.com/influxdata/influxdb/client/v2"
	"reflect"
	"time"
)

type InfluxDB struct {
	url    string
	db     string
	user   string
	pwd    string
	client client.Client

	log bool
}

func NewInfluxDB(url, user, pwd, db string) *InfluxDB {
	obj := &InfluxDB{
		url:  url,
		user: user,
		pwd:  pwd,
		db:   db,
	}

	return obj
}

// 连接服务器
func (obj *InfluxDB) Connect() error {
	// Create a new HTTPClient
	c, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     obj.url,
		Username: obj.user,
		Password: obj.pwd,
	})
	if err != nil {
		return err
	}

	obj.client = c
	logger.Info("connected influx db(%v)", obj.url)
	return nil
}

// 断开连接
func (obj *InfluxDB) DisConnect() {
	logger.Info("close influx db connection")

	if obj.client != nil {
		obj.client.Close()
	}
}

// 显示日志
func (obj *InfluxDB) ShowDBLog(log bool) {
	obj.log = log
}

// 查询
func (obj *InfluxDB) Find(cmd string, out interface{}) error {
	if obj.log {
		logger.Debug("influx: %s", cmd)
	}

	q := client.Query{
		Command:  cmd,
		Database: obj.db,
	}

	response, err := obj.client.Query(q)
	if err != nil {
		return err
	}

	if response.Error() != nil {
		return response.Error()
	}

	// 数据解析
	var data []map[string]interface{}
	for _, r := range response.Results[0].Series {
		for _, v := range r.Values {
			item := make(map[string]interface{})
			for i, c := range r.Columns {
				item[c] = v[i]
			}
			for tK, tV := range r.Tags {
				item[tK] = tV
			}
			data = append(data, item)
		}
	}

	// 查询结果为空
	//if len(data) == 0 {
	//	return errors.New("not found")
	//}

	// 输出参数类型
	t := reflect.TypeOf(out)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	var b []byte
	if t.Kind() == reflect.Slice {
		b, err = json.Marshal(data)
	} else if len(data) != 0 {
		b, err = json.Marshal(data[0])
	} else {
		return nil
	}

	if err != nil {
		return err
	}

	return json.Unmarshal(b, out)
}

func (obj *InfluxDB) GetById(table string, id string, out interface{}) error {
	cmd := fmt.Sprintf("select * from %s where Id = '%s'", table, id)

	if obj.log {
		logger.Debug("influx: %s", cmd)
	}

	return obj.Find(cmd, out)
}

/***
** purpose：
**   数据格式化
**
** params：
**   tag: tag列表
**   field: field列表
**   fieldType: 字段类型
**   data: 带格式化数据（结构体对象）
**
** return:
**   error: 错误码
***/
func parseData(tag, field []string, fieldType map[string]string, data reflect.Value) (time.Time, map[string]string, map[string]interface{}) {
	vTag := make(map[string]string)
	vField := make(map[string]interface{})

	// 指针对象
	if data.Kind() == reflect.Ptr {
		data = data.Elem()
	}

	// 普通字段
	for _, key := range field {
		v := data.FieldByName(key)

		// 空值判断
		if t := v.Kind(); ((t == reflect.Interface || t == reflect.Ptr) && v.IsNil()) ||
			((t >= reflect.Int && t <= reflect.Int64) && v.Int() == 0) ||
			((t >= reflect.Uint && t <= reflect.Uint64) && v.Uint() == 0) {
			continue
		} else if t == reflect.Slice {
			if v.IsNil() {
				continue
			}

			//vField[key] = fmt.Sprintf("%s", v)
			vField[key] = v.Interface()
		} else {
			vField[key] = v.Interface()
		}
	}

	// tag
	for _, key := range tag {
		v := data.FieldByName(key)
		vTag[key] = fmt.Sprintf("%v", v)
	}

	// 时间
	var t time.Time
	if v := data.FieldByName("Time"); v.IsValid() {
		t = v.Interface().(time.Time)
	} else {
		t = time.Now()
	}

	return t, vTag, vField
}

/***
** purpose：
**   批量插入
**
** params：
**   table: 表名
**   t: 数据结构
**   param: 数据项
**
** return:
**   error: 错误码
***/
func (obj *InfluxDB) Insert(table string, t interface{}, param interface{}) error {
	data := reflect.ValueOf(param)

	// 字段名/类型
	fieldTypes := make(map[string]string)
	// 结构体字段名
	var fields []string
	var tags []string

	// 字段解析
	{
		t1 := reflect.TypeOf(t)
		if t1.Kind() == reflect.Ptr {
			t1 = t1.Elem()
		}

		fieldNum := t1.NumField()
		for i := 0; i < fieldNum; i++ {
			tag := t1.Field(i).Tag.Get("influx")
			if tag == "-" { // 无效字段
				continue
			}

			// 判断是否是嵌套结构
			if sField := t1.Field(i).Type; sField.Kind() == reflect.Struct {
				sNum := sField.NumField()
				for j := 0; j < sNum; j++ {
					if sTag := sField.Field(j).Tag.Get("influx"); sTag == "-" { // 无效字段
						continue
					} else if sTag == "tag:true" { // tag字段
						tags = append(tags, sField.Field(j).Name)
					} else {
						fields = append(fields, sField.Field(j).Name)
					}
				}
			} else {
				if tag == "tag:true" { // tag字段
					tags = append(tags, t1.Field(i).Name)
				} else {
					fields = append(fields, t1.Field(i).Name)
				}
			}

			//fieldTypes[t1.Field(i).Name] = t1.Field(i).Type.String()
		}
	}

	bp, _ := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  obj.db,
		Precision: "us",
	})

	// 切片参数
	if data.Kind() == reflect.Slice {
		for i := 0; i < data.Len(); i++ {
			vTime, vTag, vField := parseData(tags, fields, fieldTypes, data.Index(i))
			point, _ := client.NewPoint(table, vTag, vField, vTime)
			bp.AddPoint(point)
		}
	} else {
		vTime, vTag, vField := parseData(tags, fields, fieldTypes, data)
		point, _ := client.NewPoint(table, vTag, vField, vTime)
		bp.AddPoint(point)
	}

	if obj.log {
		logger.Debug("influx: %v", bp)
	}

	if err := obj.client.Write(bp); err != nil {
		logger.Error(err.Error())
		return err
	}

	return nil
}
