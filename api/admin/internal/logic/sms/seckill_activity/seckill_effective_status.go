package seckill_activity

import "time"

// computeSeckillEffectiveStatus 根据秒杀活动 status、isEnabled 和 endTime 计算综合生效状态
// status: 0-上线, 1-下线; isEnabled: 0/1
func computeSeckillEffectiveStatus(status int32, isEnabled int32, endTime string, now time.Time) string {
	if isEnabled != 1 {
		return "未启用"
	}
	switch status {
	case 0:
		t, err := time.ParseInLocation("2006-01-02 15:04:05", endTime, time.Local)
		if err == nil && !t.After(now) {
			return "已过期"
		}
		return "已上线"
	case 1:
		return "已下线"
	default:
		return "未知"
	}
}
