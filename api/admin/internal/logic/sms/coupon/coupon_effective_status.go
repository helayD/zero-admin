package coupon

import (
	"fmt"
	"time"
)

// computeCouponEffectiveStatus 根据优惠券 status 和 endTime 计算综合生效状态
// status: 0-未开始, 1-进行中, 2-已结束, 3-已取消
func computeCouponEffectiveStatus(status int32, endTime string, now time.Time) string {
	switch status {
	case 0:
		return "未开始"
	case 1:
		t, err := time.ParseInLocation("2006-01-02 15:04:05", endTime, time.Local)
		if err == nil && !t.After(now) {
			return "已过期"
		}
		return "生效中"
	case 2:
		return "已结束"
	case 3:
		return "已取消"
	default:
		return "未知"
	}
}

// computeCouponScopeSummary 根据 scopeType 和 scopeCount 构建适用范围摘要
// scopeType: 0-全场, 1-分类, 2-商品
func computeCouponScopeSummary(scopeType int32, scopeCount int64) string {
	if scopeCount == 0 {
		return "未配置适用范围"
	}
	switch scopeType {
	case 0:
		return "全场通用"
	case 1:
		return fmt.Sprintf("指定分类(%d)", scopeCount)
	case 2:
		return fmt.Sprintf("指定商品(%d)", scopeCount)
	default:
		return "未配置适用范围"
	}
}
