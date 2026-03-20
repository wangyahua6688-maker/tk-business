package home

import "time"

// homeNowInEast8 返回首页模块统一使用的东八区当前时间。
func homeNowInEast8() time.Time {
	return time.Now().In(homeLocationEast8())
}

// homeLocationEast8 返回首页模块使用的固定时区（东八区）。
func homeLocationEast8() *time.Location {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err == nil {
		return loc
	}
	return time.FixedZone("UTC+8", 8*3600)
}
