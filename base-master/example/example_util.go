package main

import (
	"fmt"
	"git.scsv.online/go/base/logger"
	"git.scsv.online/go/base/util"
	"strings"
	"strconv"
)

func main() {

	//test_path()
	//test_buffer()
	//test_bcd()
	n, _ := strconv.ParseInt("da", 16, 32)
	logger.Debug("%d", n)
}

func test_path() {
	ss := util.URLUnescape("rtsp%3A//admin%3A12345%40192.168.1.79%3A554/Streaming/Channels/101%3Ftransportmode=unicast%26profile%3DProfile_101")
	logger.Debug(ss)
}

func test_buffer() {
	var n []byte
	buf := &util.Buffer{}

	a := []byte{0x0, 0x1, 0x2}
	b := []byte{0x3, 0x4, 0x5, 0x6, 0x7}

	logger.Debug("b[%d] cap: %d, len : %d", &b[0], cap(b), len(b))
	b = b[1:]
	logger.Debug("b[%d] cap: %d, len : %d", &b[0], cap(b), len(b))

	t := []int{1, 2, 3}
	logger.Debug(strings.Replace(strings.Trim(fmt.Sprint(t), "[]"), " ", ",", -1))

	logger.Debug(fmt.Sprint(t))
	logger.Debug(strings.Trim(fmt.Sprint(t), "[]"))

	n = append(n, a...)
	logger.Debug("n : %v", n)

	buf.Add(a)
	logger.Debug("len %d, cap %d", buf.Len(), buf.Cap())

	buf.Add(b)
	logger.Debug("len %d, cap %d", buf.Len(), buf.Cap())

	buf.Clear()
	logger.Debug("len %d, cap %d", buf.Len(), buf.Cap())

	buf.Add(b[:2])
	logger.Debug("len %d, cap %d, %v", buf.Len(), buf.Cap(), buf.Data())
}

func test_bcd() {
	bcd := []byte{0x01, 0x89, 0x69, 0x18, 0x62, 0x42}
	bcdstring := util.BCD2String(bcd)

	logger.Debug(bcdstring)

	logger.Debug(util.BCDMobile(bcd))

	bcd = util.String2BCD(bcdstring)
	logger.Debug("%02x", bcd)
}
