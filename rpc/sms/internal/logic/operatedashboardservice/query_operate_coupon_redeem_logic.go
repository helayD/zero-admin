package operatedashboardservicelogic

import (
	"context"
	"database/sql"
	"time"

	"github.com/feihua/zero-admin/pkg/operatefunnel"
	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	logiccommon "github.com/feihua/zero-admin/rpc/sms/internal/logic/common"
	"github.com/feihua/zero-admin/rpc/sms/internal/svc"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type QueryOperateCouponRedeemLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewQueryOperateCouponRedeemLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryOperateCouponRedeemLogic {
	return &QueryOperateCouponRedeemLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *QueryOperateCouponRedeemLogic) QueryOperateCouponRedeem(in *smsclient.QueryOperateCouponRedeemReq) (*smsclient.QueryOperateCouponRedeemResp, error) {
	scope, err := logiccommon.NormalizeProtoScope(in.Scope)
	if err != nil {
		return nil, err
	}

	startTime, endTime, err := operatefunnel.ParseTimeRange(in.StartTime, in.EndTime)
	if err != nil {
		return nil, err
	}
	bucket := operatefunnel.NormalizeBucket(in.Bucket, startTime, endTime)
	windows := operatefunnel.BuildBucketWindows(startTime, endTime, bucket)

	rows := make([]couponRedeemBucketRow, 0)
	query := l.baseCouponRedeemQuery(scope, in, startTime, endTime).
		Select(operateBucketExpr("cr.use_time", bucket) + " AS bucket_start, COUNT(*) AS coupon_redeem")

	if err = query.Group("bucket_start").Order("bucket_start ASC").Scan(&rows).Error; err != nil {
		return nil, err
	}

	countByBucket := make(map[string]int64, len(rows))
	for _, row := range rows {
		countByBucket[row.BucketStart] = row.CouponRedeem
	}

	trackingAt, partialMetrics, err := l.queryCouponTrackingState(scope, startTime, in)
	if err != nil {
		return nil, err
	}

	var (
		total   int64
		buckets = make([]*smsclient.OperateCouponRedeemBucketPoint, 0, len(windows))
	)
	for _, window := range windows {
		key := operatefunnel.FormatDateTime(window.Start)
		count := countByBucket[key]
		total += count
		buckets = append(buckets, &smsclient.OperateCouponRedeemBucketPoint{
			BucketLabel:  window.Label,
			BucketStart:  key,
			BucketEnd:    operatefunnel.FormatDateTime(window.End),
			CouponRedeem: count,
		})
	}

	return &smsclient.QueryOperateCouponRedeemResp{
		Buckets:           buckets,
		TotalCouponRedeem: total,
		TrackingStartedAt: nullTimeString(trackingAt),
		PartialMetrics:    partialMetrics,
	}, nil
}

func (l *QueryOperateCouponRedeemLogic) baseCouponRedeemQuery(scope pkgscope.GovernanceScope, in *smsclient.QueryOperateCouponRedeemReq, startTime, endTime time.Time) *gorm.DB {
	query := l.couponRedeemScopedQuery(scope).
		Where("cr.use_time >= ? AND cr.use_time < ?", startTime, endTime)

	return applyCouponRedeemFilters(query, in)
}

func (l *QueryOperateCouponRedeemLogic) queryCouponTrackingState(scope pkgscope.GovernanceScope, startTime time.Time, in *smsclient.QueryOperateCouponRedeemReq) (sql.NullTime, []string, error) {
	var value sql.NullTime
	partialMetrics := make([]string, 0, 1)
	activityType, ok := operatefunnel.OptionalActivityType(in.ActivityType)
	if !ok || activityType == operatefunnel.ActivityNone {
		return value, partialMetrics, nil
	}

	query := applyCouponRedeemFilters(
		l.couponRedeemScopedQuery(scope).Select("MIN(cr.use_time)"),
		in,
	)
	if err := query.Scan(&value).Error; err != nil {
		return sql.NullTime{}, nil, err
	}
	if value.Valid && startTime.Before(value.Time) {
		partialMetrics = append(partialMetrics, operatefunnel.EventCouponRedeem)
	}
	return value, partialMetrics, nil
}

func (l *QueryOperateCouponRedeemLogic) couponRedeemScopedQuery(scope pkgscope.GovernanceScope) *gorm.DB {
	return l.svcCtx.DB.WithContext(l.ctx).
		Table("sms_coupon_record cr").
		Joins("LEFT JOIN oms_order_main o ON o.id = cr.order_id AND o.is_deleted = 0").
		Joins("LEFT JOIN sms_coupon c ON c.id = cr.coupon_id AND c.is_deleted = 0").
		Where("cr.status = 1 AND cr.use_time IS NOT NULL").
		Where("(o.id IS NOT NULL AND o.platform_id = ? AND o.tenant_id = ? AND o.merchant_id = ?) OR (o.id IS NULL AND c.platform_id = ? AND c.tenant_id = ? AND c.merchant_id = ?)",
			scope.PlatformID, scope.TenantID, scope.MerchantID,
			scope.PlatformID, scope.TenantID, scope.MerchantID,
		)
}

func applyCouponRedeemFilters(query *gorm.DB, in *smsclient.QueryOperateCouponRedeemReq) *gorm.DB {
	if channel, ok := operatefunnel.OptionalChannel(in.Channel); ok {
		query = query.Where(couponRedeemChannelCaseSQL()+" = ?", channel)
	}

	if activityType, ok := operatefunnel.OptionalActivityType(in.ActivityType); ok {
		query = query.Where(couponRedeemActivityTypeCaseSQL()+" = ?", activityType)
		if in.ActivityId > 0 {
			query = query.Where(couponRedeemActivityIDCaseSQL()+" = ?", in.ActivityId)
		}
	}

	return query
}

func couponRedeemChannelCaseSQL() string {
	return "CASE o.source_type WHEN 1 THEN 'app' WHEN 2 THEN 'pc' WHEN 3 THEN 'mini_program' ELSE 'unknown' END"
}

func couponRedeemActivityTypeCaseSQL() string {
	return "CASE WHEN COALESCE(NULLIF(o.activity_type, ''), 'none') <> 'none' THEN o.activity_type WHEN c.id IS NOT NULL THEN 'coupon' ELSE 'none' END"
}

func couponRedeemActivityIDCaseSQL() string {
	return "CASE WHEN COALESCE(NULLIF(o.activity_type, ''), 'none') <> 'none' THEN o.activity_id WHEN c.id IS NOT NULL THEN c.id ELSE 0 END"
}
