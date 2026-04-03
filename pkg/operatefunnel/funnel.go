package operatefunnel

import (
	"fmt"
	"strings"
	"time"
)

const (
	EventExposure     = "exposure"
	EventClick        = "click"
	EventAddCart      = "add_cart"
	EventOrderCreated = "order_created"
	EventPaySuccess   = "pay_success"
	EventCouponRedeem = "coupon_redeem"
)

const (
	ChannelApp         = "app"
	ChannelPC          = "pc"
	ChannelH5          = "h5"
	ChannelMiniProgram = "mini_program"
	ChannelUnknown     = "unknown"
)

const (
	ActivityNone           = "none"
	ActivityHomeAdvertise  = "home_advertise"
	ActivityCoupon         = "coupon"
	ActivitySeckillActivity = "seckill_activity"
)

const (
	BucketDay  = "day"
	BucketHour = "hour"
)

const (
	TimeLayoutDateTime = "2006-01-02 15:04:05"
	TimeLayoutDate     = "2006-01-02"
)

var shanghaiLocation = mustLoadShanghai()

type BucketWindow struct {
	Start time.Time
	End   time.Time
	Label string
}

func mustLoadShanghai() *time.Location {
	location, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		return time.FixedZone("Asia/Shanghai", 8*3600)
	}
	return location
}

func Location() *time.Location {
	return shanghaiLocation
}

func FunnelStages() []string {
	return []string{
		EventExposure,
		EventClick,
		EventAddCart,
		EventOrderCreated,
		EventPaySuccess,
		EventCouponRedeem,
	}
}

func ParseTimeRange(startText, endText string) (time.Time, time.Time, error) {
	start, err := ParseTime(startText)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("开始时间格式非法: %w", err)
	}

	end, err := ParseTime(endText)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("结束时间格式非法: %w", err)
	}

	if !start.Before(end) {
		return time.Time{}, time.Time{}, fmt.Errorf("开始时间必须早于结束时间")
	}

	return start, end, nil
}

func ParseTime(text string) (time.Time, error) {
	trimmed := strings.TrimSpace(text)
	for _, layout := range []string{time.RFC3339, TimeLayoutDateTime, TimeLayoutDate} {
		value, err := time.ParseInLocation(layout, trimmed, Location())
		if err == nil {
			return value.In(Location()), nil
		}
	}
	return time.Time{}, fmt.Errorf("unsupported time layout: %s", text)
}

func NormalizeBucket(bucket string, start, end time.Time) string {
	switch strings.TrimSpace(bucket) {
	case BucketHour:
		if end.Sub(start) <= 48*time.Hour {
			return BucketHour
		}
	case BucketDay:
		return BucketDay
	}
	return BucketDay
}

func BucketStart(value time.Time, bucket string) time.Time {
	local := value.In(Location())
	switch bucket {
	case BucketHour:
		return time.Date(local.Year(), local.Month(), local.Day(), local.Hour(), 0, 0, 0, Location())
	default:
		return time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, Location())
	}
}

func NextBucketStart(value time.Time, bucket string) time.Time {
	switch bucket {
	case BucketHour:
		return BucketStart(value, bucket).Add(time.Hour)
	default:
		return BucketStart(value, bucket).AddDate(0, 0, 1)
	}
}

func BucketLabel(value time.Time, bucket string) string {
	switch bucket {
	case BucketHour:
		return BucketStart(value, bucket).Format("2006-01-02 15:00")
	default:
		return BucketStart(value, bucket).Format(TimeLayoutDate)
	}
}

func IterateBuckets(start, end time.Time, bucket string, fn func(bucketStart, bucketEnd time.Time)) {
	for current := BucketStart(start, bucket); current.Before(end); current = NextBucketStart(current, bucket) {
		next := NextBucketStart(current, bucket)
		fn(current, next)
	}
}

func NormalizeChannel(channel string) string {
	switch strings.TrimSpace(channel) {
	case ChannelApp, ChannelPC, ChannelH5, ChannelMiniProgram:
		return strings.TrimSpace(channel)
	default:
		return ChannelUnknown
	}
}

func NormalizeActivityType(activityType string) string {
	switch strings.TrimSpace(activityType) {
	case ActivityHomeAdvertise, ActivityCoupon, ActivitySeckillActivity:
		return strings.TrimSpace(activityType)
	default:
		return ActivityNone
	}
}

func OptionalChannel(channel string) (string, bool) {
	trimmed := strings.TrimSpace(channel)
	if trimmed == "" {
		return "", false
	}
	return NormalizeChannel(trimmed), true
}

func OptionalActivityType(activityType string) (string, bool) {
	trimmed := strings.TrimSpace(activityType)
	if trimmed == "" {
		return "", false
	}
	return NormalizeActivityType(trimmed), true
}

func OrderChannelFromSource(sourceType int32) string {
	switch sourceType {
	case 1:
		return ChannelApp
	case 2:
		return ChannelPC
	case 3:
		return ChannelMiniProgram
	default:
		return ChannelUnknown
	}
}

func CartChannelFromSource(source int32) string {
	switch source {
	case 1:
		return ChannelPC
	case 2:
		return ChannelH5
	case 3:
		return ChannelMiniProgram
	case 4:
		return ChannelApp
	default:
		return ChannelUnknown
	}
}

func SafeRate(numerator, denominator int64) float64 {
	if denominator <= 0 {
		return 0
	}
	return float64(numerator) / float64(denominator)
}

func FormatDateTime(value time.Time) string {
	return value.In(Location()).Format(TimeLayoutDateTime)
}

func BuildBucketWindows(start, end time.Time, bucket string) []BucketWindow {
	windows := make([]BucketWindow, 0)
	IterateBuckets(start, end, bucket, func(bucketStart, bucketEnd time.Time) {
		windows = append(windows, BucketWindow{
			Start: bucketStart,
			End:   bucketEnd,
			Label: BucketLabel(bucketStart, bucket),
		})
	})
	return windows
}
