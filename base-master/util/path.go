package util

import (
	"fmt"
	"os"
	"path/filepath"
)

/*
获取程序运行路径
*/
func GetCurrentDirectory() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		fmt.Println(err)
	}
	//return strings.Replace(dir, "\\", "/", -1)
	return dir
}

/*
替换URL特殊字符
*/
var mapURLDecode = map[string]byte{
	"%20": ' ',
	"%22": '"',
	"%23": '#',
	"%25": '%',
	"%26": '&',
	"%28": '(',
	"%29": ')',
	"%2B": '+',
	"%2C": ',',
	"%2F": '/',
	"%3A": ':',
	"%3B": ';',
	"%3C": '<',
	"%3D": '=',
	"%3E": '>',
	"%3F": '?',
	"%40": '@',
	"%5C": '\\',
	"%7C": '|',
}

func URLUnescape(url string) string {
	var ss string
	var cc byte

	l := len(url)
	ret := []byte{}
	i := 0
	j := 0

	for {
		cc = url[i]
		if cc != '%' {
			i++
			if i >= l {
				break
			}
			continue
		}

		if i+2 >= l {
			break
		}

		ss = url[i : i+3]
		cc = mapURLDecode[ss]
		ret = append(ret, url[:i]...)
		ret = append(ret, cc)
		url = url[i+3:]
		l = len(url)
		i = 0
		j = 0
	}

	if i > j {
		ret = append(ret, url[j:i]...)
	}

	return string(ret)
}
