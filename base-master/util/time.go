package util

import (
	"time"
)

const timeFormat = "2006-01-02 15:04:05"

// 字符串时间格式化
func ParseTime(t string) (time.Time, error) {
	return time.Parse(timeFormat, t)
}
