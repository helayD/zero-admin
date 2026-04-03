package operatedashboardservicelogic

import (
	"database/sql"

	"github.com/feihua/zero-admin/pkg/operatefunnel"
	pkgscope "github.com/feihua/zero-admin/pkg/scope"
)

type trafficBucketRow struct {
	BucketStart string `gorm:"column:bucket_start"`
	Exposure    int64  `gorm:"column:exposure"`
	Click       int64  `gorm:"column:click"`
}

type couponRedeemBucketRow struct {
	BucketStart   string `gorm:"column:bucket_start"`
	CouponRedeem  int64  `gorm:"column:coupon_redeem"`
}

type activityOptionRow struct {
	ActivityType  string `gorm:"column:activity_type"`
	ActivityID    int64  `gorm:"column:activity_id"`
	ActivityName  string `gorm:"column:activity_name"`
	ActivityLabel string `gorm:"column:activity_label"`
}

func operateBucketExpr(column, bucket string) string {
	switch bucket {
	case operatefunnel.BucketHour:
		return "DATE_FORMAT(" + column + ", '%Y-%m-%d %H:00:00')"
	default:
		return "DATE_FORMAT(" + column + ", '%Y-%m-%d 00:00:00')"
	}
}

func nullTimeString(value sql.NullTime) string {
	if !value.Valid {
		return ""
	}
	return operatefunnel.FormatDateTime(value.Time)
}

func scopeArgs(scope pkgscope.GovernanceScope) []interface{} {
	return []interface{}{scope.PlatformID, scope.TenantID, scope.MerchantID}
}
