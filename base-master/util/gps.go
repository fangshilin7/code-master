package util

import (
	"math"
)

const PI = 3.14159265358979324

func transformLat(x, y float64) float64 {
	ret := -100.0 + 2.0*x + 3.0*y + 0.2*y*y + 0.1*x*y + 0.2*math.Sqrt(math.Abs(x))
	ret += (20.0*math.Sin(6.0*x*PI) + 20.0*math.Sin(2.0*x*PI)) * 2.0 / 3.0
	ret += (20.0*math.Sin(y*PI) + 40.0*math.Sin(y/3.0*PI)) * 2.0 / 3.0
	ret += (160.0*math.Sin(y/12.0*PI) + 320*math.Sin(y*PI/30.0)) * 2.0 / 3.0

	return ret
}
func transformLon(x, y float64) float64 {
	ret := 300.0 + x + 2.0*y + 0.1*x*x + 0.1*x*y + 0.1*math.Sqrt(math.Abs(x))
	ret += (20.0*math.Sin(6.0*x*PI) + 20.0*math.Sin(2.0*x*PI)) * 2.0 / 3.0
	ret += (20.0*math.Sin(x*PI) + 40.0*math.Sin(x/3.0*PI)) * 2.0 / 3.0
	ret += (150.0*math.Sin(x/12.0*PI) + 300.0*math.Sin(x/30.0*PI)) * 2.0 / 3.0

	return ret
}

func outOfChina(lat, lng float64) bool {
	if lng < 72.004 || lng > 137.8347 {
		return true
	}
	if lat < 0.8293 || lat > 55.8271 {
		return true
	}
	return false
}

func delta(lat, lng float64) (float64, float64) {
	// Krasovsky 1940
	//
	// a = 6378245.0, 1/f = 298.3
	// b = a * (1 - f)
	// ee = (a^2 - b^2) / a^2;
	var a = 6378245.0               //  a: 卫星椭球坐标投影到平面地图坐标系的投影因子。
	var ee = 0.00669342162296594323 //  ee: 椭球的偏心率。
	var dLat = transformLat(lng-105.0, lat-35.0)
	var dLon = transformLon(lng-105.0, lat-35.0)

	var radLat = lat / 180.0 * PI
	var magic = math.Sin(radLat)
	magic = 1 - ee*magic*magic
	var sqrtMagic = math.Sqrt(magic)
	dLat = (dLat * 180.0) / ((a * (1 - ee)) / (magic * sqrtMagic) * PI)
	dLon = (dLon * 180.0) / (a / sqrtMagic * math.Cos(radLat) * PI)
	// return { 'lat': dLat, 'lng': dLon }
	return dLon, dLat
}

// GPS---高德(纬度，经度)
func GPSToGaoDe(wgsLat, wgsLon float64) (float64, float64) {
	if outOfChina(wgsLat, wgsLon) {
		return wgsLat, wgsLon
	}

	lon, lat := delta(wgsLat, wgsLon)
	// return { 'lat': wgsLat + d.lat, 'lng': wgsLon + d.lng }
	return wgsLat + lat, wgsLon + lon
}
