package util

import (
	"encoding/hex"
)

//BCD码转String
func BCD2String(data []byte) string {
	return hex.EncodeToString(data)
}

//string转BCD码
func String2BCD(s string) []byte {
	b, _ := hex.DecodeString(s)
	return b
}

//BCD码手机号
func BCDMobile(data []byte) string {
	s := BCD2String(data)
	if len(s) > 0 {
		s = s[1:]
	}
	return s
}
