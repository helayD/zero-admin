package operatefunnel

import (
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	pkgscope "github.com/feihua/zero-admin/pkg/scope"
)

const (
	RepeatPurchaseWindowDays = 180

	RepeatPurchaseMetricPaidBuyerCount   = "paidBuyerCount"
	RepeatPurchaseMetricRepeatBuyerCount = "repeatBuyerCount"
	RepeatPurchaseMetricRepeatRate       = "repeatRate"
	RepeatPurchaseMetricRepeatOrderCount = "repeatOrderCount"
	RepeatPurchaseMetricRepeatGmv        = "repeatGmv"
	RepeatPurchaseMetricAvgDaysToRepeat  = "avgDaysToRepeat"
)

var repeatPurchaseMetricKeys = []string{
	RepeatPurchaseMetricPaidBuyerCount,
	RepeatPurchaseMetricRepeatBuyerCount,
	RepeatPurchaseMetricRepeatRate,
	RepeatPurchaseMetricRepeatOrderCount,
	RepeatPurchaseMetricRepeatGmv,
	RepeatPurchaseMetricAvgDaysToRepeat,
}

type RepeatPurchaseMetricDefinition struct {
	Key   string
	Label string
	Unit  string
}

type RepeatPurchaseDetailColumn struct {
	Key   string
	Label string
}

func RepeatPurchaseLookbackWindow(payTime time.Time) (time.Time, time.Time, error) {
	if payTime.IsZero() {
		return time.Time{}, time.Time{}, fmt.Errorf("支付时间不能为空")
	}
	end := payTime.In(Location())
	start := end.AddDate(0, 0, -RepeatPurchaseWindowDays)
	return start, end, nil
}

func IsRepeatPurchaseEffectiveOrder(isDeleted int32, payTime time.Time, orderStatus int32) bool {
	if isDeleted != 0 || payTime.IsZero() {
		return false
	}

	switch orderStatus {
	case 2, 3, 4, 7:
		return true
	case 5, 6:
		return false
	default:
		return false
	}
}

func RepeatPurchaseScopeOrderColumns(scope pkgscope.GovernanceScope) ([]string, error) {
	switch strings.TrimSpace(scope.ScopeType) {
	case pkgscope.SubjectTypePlatform:
		if scope.PlatformID <= 0 {
			return nil, fmt.Errorf("平台级复购分析必须提供 platformId")
		}
		return []string{"platform_id"}, nil
	case pkgscope.SubjectTypeTenant:
		if scope.PlatformID <= 0 || scope.TenantID <= 0 {
			return nil, fmt.Errorf("租户级复购分析必须提供 platformId 和 tenantId")
		}
		return []string{"platform_id", "tenant_id"}, nil
	case pkgscope.SubjectTypeMerchant:
		if scope.PlatformID <= 0 || scope.TenantID <= 0 || scope.MerchantID <= 0 {
			return nil, fmt.Errorf("商户级复购分析必须提供 platformId、tenantId 和 merchantId")
		}
		return []string{"platform_id", "tenant_id", "merchant_id"}, nil
	default:
		return nil, fmt.Errorf("不支持的复购分析作用域：%s", scope.ScopeType)
	}
}

func RepeatPurchaseOverviewDefinitions() []RepeatPurchaseMetricDefinition {
	return []RepeatPurchaseMetricDefinition{
		{Key: RepeatPurchaseMetricPaidBuyerCount, Label: "支付买家数", Unit: "人"},
		{Key: RepeatPurchaseMetricRepeatBuyerCount, Label: "复购买家数", Unit: "人"},
		{Key: RepeatPurchaseMetricRepeatRate, Label: "复购率", Unit: "%"},
		{Key: RepeatPurchaseMetricRepeatOrderCount, Label: "复购订单数", Unit: "单"},
		{Key: RepeatPurchaseMetricRepeatGmv, Label: "复购GMV", Unit: "元"},
		{Key: RepeatPurchaseMetricAvgDaysToRepeat, Label: "平均复购天数", Unit: "天"},
	}
}

func RepeatPurchaseDetailColumns() []RepeatPurchaseDetailColumn {
	return []RepeatPurchaseDetailColumn{
		{Key: "memberId", Label: "会员ID"},
		{Key: "nicknameMasked", Label: "昵称"},
		{Key: "mobileMasked", Label: "手机号"},
		{Key: "firstValidPayTime", Label: "首次有效支付时间"},
		{Key: "latestRepeatPayTime", Label: "最近复购支付时间"},
		{Key: "repeatOrderCount", Label: "复购订单数"},
		{Key: "repeatGmv", Label: "复购GMV"},
		{Key: "latestChannel", Label: "最近渠道"},
		{Key: "latestActivityType", Label: "最近活动类型"},
		{Key: "latestActivityId", Label: "最近活动ID"},
		{Key: "platformId", Label: "平台ID"},
		{Key: "tenantId", Label: "租户ID"},
		{Key: "merchantId", Label: "商户ID"},
	}
}

func RepeatPurchaseStableSortClause() string {
	return "latest_repeat_pay_time DESC, member_id ASC"
}

func RepeatPurchasePartialMetrics(trackingStartedAt, queryStart time.Time, activityType string, activityID int64) []string {
	if trackingStartedAt.IsZero() || !queryStart.Before(trackingStartedAt) {
		return nil
	}
	if strings.TrimSpace(activityType) == "" && activityID <= 0 {
		return nil
	}
	result := make([]string, 0, len(repeatPurchaseMetricKeys))
	result = append(result, repeatPurchaseMetricKeys...)
	return result
}

func MaskRepeatPurchaseMobile(mobile string) string {
	trimmed := strings.TrimSpace(mobile)
	if len(trimmed) < 7 {
		return trimmed
	}
	return trimmed[:3] + "****" + trimmed[len(trimmed)-4:]
}

func MaskRepeatPurchaseNickname(nickname string) string {
	trimmed := strings.TrimSpace(nickname)
	runeCount := utf8.RuneCountInString(trimmed)
	if runeCount <= 1 {
		return trimmed
	}
	if runeCount == 2 {
		runes := []rune(trimmed)
		return string(runes[0]) + "*"
	}
	runes := []rune(trimmed)
	return string(runes[0]) + "**" + string(runes[len(runes)-1])
}
