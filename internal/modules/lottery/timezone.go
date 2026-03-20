package lottery

import "time"

// lotteryNowInEast8 返回当前东八区时间，作为开奖业务统一时区基准。
func lotteryNowInEast8() time.Time {
	return time.Now().In(lotteryLocationEast8())
}

// lotteryLocationEast8 返回开奖业务使用的固定时区（东八区）。
func lotteryLocationEast8() *time.Location {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err == nil {
		return loc
	}
	return time.FixedZone("UTC+8", 8*3600)
}
