package util

import (
	"git.scsv.online/go/base/logger"
	"regexp"
)

func IsMobile(mobile string) bool {
	mobileReg, err := regexp.Compile("^1[0-9]{10}$")
	if err != nil {
		logger.Error("get mobile regex object error")
		return false
	}
	return mobileReg.MatchString(mobile)
}

func IsIP(ip string) bool {
	ipReg, err := regexp.Compile("^(\\d|[1-9]\\d|1\\d{2}|2[0-5][0-5])\\.(\\d|[1-9]\\d|1\\d{2}|2[0-5][0-5])\\.(\\d|[1-9]\\d|1\\d{2}|2[0-5][0-5])\\.(\\d|[1-9]\\d|1\\d{2}|2[0-5][0-5])$")
	if err != nil {
		logger.Error("get ip regex object error")
		return false
	}

	return ipReg.MatchString(ip)
}

func IsPort(port string) bool {
	portReg, err := regexp.Compile("^([0-9]|[1-9]\\d{1,3}|[1-5]\\d{4}|6[0-5]{2}[0-3][0-5])$")
	if err != nil {
		logger.Error("get port regex object error")
		return false
	}
	return portReg.MatchString(port)
}

func IsQQ(qq string) bool {
	qqReg, err := regexp.Compile("^[1-9][0-9]{4,9}$")
	if err != nil {
		logger.Error("get qq regex object error")
		return false
	}
	return qqReg.MatchString(qq)
}

func IsEmail(email string) bool {
	emailReg, err := regexp.Compile("^[a-z0-9]+([._\\-]*[a-z0-9])*@([a-z0-9]+[-a-z0-9]*[a-z0-9]+.){1,63}[a-z0-9]+$")
	if err != nil {
		logger.Error("get email regex object error")
		return false
	}
	return emailReg.MatchString(email)
}