package util

import (
	"crypto/md5"
	"fmt"
	"time"
)

func RC() string {
	m := md5.New()
	m.Write([]byte(fmt.Sprintf("%d", time.Now().UnixNano())))
	return fmt.Sprintf("%x", m.Sum(nil))[:12]
}

func RCRand(key string) string {
	m := md5.New()
	m.Write([]byte(fmt.Sprintf("%d%s", time.Now().UnixNano(), key)))
	return fmt.Sprintf("%x", m.Sum(nil))[:12]
}
