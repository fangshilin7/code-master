package db

import (
	"fmt"
	"git.scsv.online/go/base/logger"
	"git.scsv.online/go/base/util"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"reflect"
	"strings"
	"time"
)

const (
	Time_Layout  = "2006-01-02 15:04:05"
	Date_Layout  = "2006-01-02"
	Month_Layout = "200601"
)

type MySqlBase struct {
	Id string `gorm:"size:32;primary_key" json:",omitempty"`
}

type MySqlDB struct {
	url    string
	client *gorm.DB
}

func NewMySqlDB(url string) *MySqlDB {
	obj := &MySqlDB{
		url: url,
	}

	return obj
}

// 连接服务器
func (obj *MySqlDB) Connect() error {
	c, err := gorm.Open("mysql", obj.url)
	if err != nil {
		return err
	}

	// 连接池
	c.DB().SetMaxIdleConns(2)
	c.DB().SetMaxOpenConns(5)

	// 全局禁用表名复数
	c.SingularTable(true)

	obj.client = c
	logger.Info("connected mysql db(%v)", obj.url)
	return nil
}

// 断开连接
func (obj *MySqlDB) DisConnect() {
	logger.Info("close mysql db connection")

	if obj.client != nil {
		obj.client.Close()
	}
}

// 数据库操作日志
func (obj *MySqlDB) ShowDBLog(show bool) {
	// 启用Logger，显示详细日志
	obj.client.LogMode(show)
}

// 自动迁移
func (obj *MySqlDB) Migrate(params ...interface{}) {
	obj.client.AutoMigrate(params...)
}

// 查询函数
// 表名通过扩展参数传递
func (obj *MySqlDB) Find(param *FindParam, out interface{}, ext ...string) error {
	c := obj.client
	var err error

	// 指定表名
	if len(ext) != 0 {
		c = c.Table(ext[0])
	}

	// 排序
	if param.Sort != "" {
		c = c.Order(param.Sort)
	}

	// 查询条件
	if param.Query != nil {
		// 多参数
		if param.Option == nil {
			c = c.Where(param.Query)
		} else if v, ok := param.Option.([]interface{}); ok {
			c = c.Where(param.Query, v...)
		} else {
			c = c.Where(param.Query, param.Option)
		}
	}

	// 输出参数类型
	t := reflect.TypeOf(out)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() == reflect.Slice {
		err = c.Find(out).Error
	} else {
		err = c.First(out).Error
	}

	return err
}

// 创建
func (obj *MySqlDB) Insert(param interface{}) error {
	return obj.client.Create(param).Error
}
func (obj *MySqlDB) InsertToTable(table string, param interface{}) error {
	return obj.client.Table(table).Create(param).Error
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
func (obj *MySqlDB) BatchInsertToTable(table string, t interface{}, param interface{}) error {
	data := reflect.ValueOf(param)

	// 不是切片参数
	if data.Kind() != reflect.Slice {
		return fmt.Errorf("param type isn't slice, %#v", param)
	}

	// 字段名/类型
	fieldTypes := make(map[string]string)
	// 结构体字段名
	var fields []string
	// 数据库字段名
	var columns []string

	// 字段解析
	t1 := reflect.TypeOf(t)
	if t1.Kind() == reflect.Ptr {
		t1 = t1.Elem()
	}

	fieldNum := t1.NumField()
	for i := 0; i < fieldNum; i++ {
		// 无效字段
		if t1.Field(i).Tag.Get("sql") == "-" {
			continue
		}

		name := t1.Field(i).Name
		fields = append(fields, name)
		columns = append(columns, util.SnakeString(name))
		fieldTypes[t1.Field(i).Name] = t1.Field(i).Type.String()
	}

	var values []string
	for i := 0; i < data.Len(); i++ {
		var value []string
		t2 := data.Index(i)
		// 指针对象
		if t2.Kind() == reflect.Ptr {
			t2 = t2.Elem()
		}

		for _, key := range fields {
			v := t2.FieldByName(key)

			if fieldTypes[key] == "time.Time" {
				value = append(value, v.Interface().(time.Time).Format(Time_Layout))
			} else if fieldTypes[key] == "*time.Time" {
				value = append(value, v.Interface().(*time.Time).Format(Time_Layout))
			} else {
				value = append(value, fmt.Sprintf("%v", v))
			}
		}

		values = append(values, fmt.Sprintf("('%s')", strings.Join(value, "', '")))
	}

	// 数据项为空
	if data.Len() == 0 {
		return nil
	}

	sql := fmt.Sprintf("insert into %s(`%s`) values%s",
		table,
		strings.ToLower(strings.Join(columns, "`, `")),
		strings.Join(values, ","))

	return obj.RawQuery(sql, nil)
}

// 根据Id查询
func (obj *MySqlDB) GetById(id string, out interface{}) bool {
	return obj.client.First(out, "id = ?", id).RecordNotFound()
}

// 更新
func (obj *MySqlDB) Update(value interface{}, doc interface{}) error {
	return obj.client.Model(value).Updates(doc).Error
}

//批量更新
func (obj *MySqlDB) BatchUpdates(table string, doc map[string]interface{}, query interface{}, param ...interface{}) error {
	return obj.client.Table(table).Where(query, param).Updates(doc).Error
}

// 删除数据
func (obj *MySqlDB) RemoveById(table string, id interface{}) error {
	return obj.client.Table(table).Delete("","id = ?", id).Error
}
func (obj *MySqlDB) RemoveByCond(value, query interface{}, param ...interface{}) error {
	return obj.client.Delete(value, query, param).Error
}

// 是否存在
func (obj *MySqlDB) Exist(table string, query interface{}, param ...interface{}) (MySqlBase, bool) {
	var out MySqlBase
	not := obj.client.Table(table).Select("id").Where(query, param...).First(&out).RecordNotFound()

	return out, not
}

// table备份
func (obj *MySqlDB) BackUpTable(table, newTable string) error {
	tx := obj.client.Begin()
	if err := tx.Exec(fmt.Sprintf("rename table %s to %s;", table, newTable)).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Exec(fmt.Sprintf("create table %s like %s;", table, newTable)).Error; err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()

	return nil
}

// 执行sql, 查询
func (obj *MySqlDB) RawQuery(sql string, out interface{}) error {
	if out == nil {
		return obj.client.Exec(sql).Error
	} else {
		return obj.client.Raw(sql).Scan(out).Error
	}
}

// 执行sql, 统计数量
func (obj *MySqlDB) RawCount(sql string) (error, uint32) {
	var count uint32

	err := obj.client.Raw(sql).Count(&count).Error

	return err, count
}
