package home_advertise

import "time"

// computeHomeAdvertiseEffectiveStatus 根据广告状态和投放时间计算综合生效状态
// status: 0-下线，1-上线
func computeHomeAdvertiseEffectiveStatus(status int32, startTime, endTime string, now time.Time) string {
	if status != 1 {
		return "已下线"
	}

	start, startErr := time.ParseInLocation("2006-01-02 15:04:05", startTime, time.Local)
	if startErr == nil && now.Before(start) {
		return "未开始"
	}

	end, endErr := time.ParseInLocation("2006-01-02 15:04:05", endTime, time.Local)
	if endErr == nil && !end.After(now) {
		return "已过期"
	}

	return "已上线"
}
