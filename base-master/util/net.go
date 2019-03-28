package util

import (
	"io"
	"net"
	"strings"
)

var localIP_ string

//获取本地IP
func GetLocalIP() string {
	if localIP_ != "" {
		return localIP_
	}

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "127.0.0.1"
	}

	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				localIP_ = ipnet.IP.String()
				break
			}
		}
	}

	if localIP_ == "" {
		localIP_ = "127.0.0.1"
	}
	return localIP_
}

//设置本地IP
func SetLocalIP(ip string) {
	localIP_ = ip
}

func GetURLProtocol(url string) string {
	i := strings.Index(url, ":")
	if i == -1 {
		return ""
	}

	protocol := strings.ToLower(url[:i])
	return protocol
}

func FullWrite(w io.Writer, buf []byte) (bytes int, err error) {
	total := len(buf)
	var offset, n int
	for {
		n, err = w.Write(buf[offset:])
		if err != nil {
			break
		}

		offset += n
		if offset >= total {
			break
		}
	}
	bytes = offset
	return
}

// 网路匹配
func IsSameNetSeg(ip1, ip2, mask string) bool {
	// 掩码默认255.255.255.0
	if mask == "" {
		mask = "255.255.255.0"
	}

	v1 := net.ParseIP(ip1)
	v2 := net.ParseIP(ip2)
	v3 := net.ParseIP(mask)

	v := net.IPNet{v2, net.IPMask(v3)}
	return v.Contains(v1)
}
