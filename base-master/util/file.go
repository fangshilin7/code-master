package util

import (
	"errors"
	"io"
	"os"
)

//拷贝文件  要拷贝的文件路径 拷贝到哪里
func Copy(source, dest string) error {
	if source == "" || dest == "" {
		return errors.New("source or dest is null")
	}

	//打开文件资源
	source_open, err := os.Open(source)
	if err != nil {
		return err
	}
	defer source_open.Close()

	//只写模式打开文件 如果文件不存在进行创建 并赋予 644的权限。详情查看linux 权限解释
	dest_open, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY, 644)
	if err != nil {
		return err
	}
	defer dest_open.Close()

	//进行数据拷贝
	_, err = io.Copy(dest_open, source_open)
	return err
}
