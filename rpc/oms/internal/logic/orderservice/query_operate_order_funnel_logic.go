package orderservicelogic

import (
	"context"
	"database/sql"
	"time"

	"github.com/feihua/zero-admin/pkg/operatefunnel"
	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	logiccommon "github.com/feihua/zero-admin/rpc/oms/internal/logic/common"
	"github.com/feihua/zero-admin/rpc/oms/internal/svc"
	"github.com/feihua/zero-admin/rpc/oms/omsclient"
	"github.com/zeromicro/go-zero/core/logx"
)

type QueryOperateOrderFunnelLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

type operateOrderBucketRow struct {
	BucketStart  string `gorm:"column:bucket_start"`
	OrderCreated int64  `gorm:"column:order_created"`
	PaySuccess   int64  `gorm:"column:pay_success"`
}

func NewQueryOperateOrderFunnelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryOperateOrderFunnelLogic {
	return &QueryOperateOrderFunnelLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查询经营漏斗订单聚合
func (l *QueryOperateOrderFunnelLogic) QueryOperateOrderFunnel(in *omsclient.QueryOperateOrderFunnelReq) (*omsclient.QueryOperateOrderFunnelResp, error) {
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

	createdRows, err := l.queryOrderCreatedBuckets(scope, in, startTime, endTime, bucket)
	if err != nil {
		return nil, err
	}
	payRows, err := l.queryPaySuccessBuckets(scope, in, startTime, endTime, bucket)
	if err != nil {
		return nil, err
	}

	createdByBucket := make(map[string]int64, len(createdRows))
	for _, row := range createdRows {
		createdByBucket[row.BucketStart] = row.OrderCreated
	}
	payByBucket := make(map[string]int64, len(payRows))
	for _, row := range payRows {
		payByBucket[row.BucketStart] = row.PaySuccess
	}

	var (
		totalCreated   int64
		totalPaid      int64
		buckets        = make([]*omsclient.OperateOrderBucketPoint, 0, len(windows))
		trackingAt     string
		partialMetrics []string
	)

	activityRequested := false
	if activityType, ok := operatefunnel.OptionalActivityType(in.ActivityType); ok {
		activityRequested = activityType != operatefunnel.ActivityNone
	}
	if activityRequested {
		createdTrack, createdOK, trackErr := l.queryOrderTrackingStart("create_time")
		if trackErr != nil {
			return nil, trackErr
		}
		payTrack, payOK, payErr := l.queryOrderTrackingStart("pay_time")
		if payErr != nil {
			return nil, payErr
		}
		if createdOK && startTime.Before(createdTrack.Time) {
			partialMetrics = append(partialMetrics, operatefunnel.EventOrderCreated)
		}
		if payOK && startTime.Before(payTrack.Time) {
			partialMetrics = append(partialMetrics, operatefunnel.EventPaySuccess)
		}
		switch {
		case createdOK && payOK:
			if createdTrack.Time.Before(payTrack.Time) {
				trackingAt = operatefunnel.FormatDateTime(createdTrack.Time)
			} else {
				trackingAt = operatefunnel.FormatDateTime(payTrack.Time)
			}
		case createdOK:
			trackingAt = operatefunnel.FormatDateTime(createdTrack.Time)
		case payOK:
			trackingAt = operatefunnel.FormatDateTime(payTrack.Time)
		}
	}

	for _, window := range windows {
		key := operatefunnel.FormatDateTime(window.Start)
		created := createdByBucket[key]
		paid := payByBucket[key]
		totalCreated += created
		totalPaid += paid
		buckets = append(buckets, &omsclient.OperateOrderBucketPoint{
			BucketLabel:  window.Label,
			BucketStart:  key,
			BucketEnd:    operatefunnel.FormatDateTime(window.End),
			OrderCreated: created,
			PaySuccess:   paid,
		})
	}

	return &omsclient.QueryOperateOrderFunnelResp{
		Buckets:           buckets,
		TotalOrderCreated: totalCreated,
		TotalPaySuccess:   totalPaid,
		TrackingStartedAt: trackingAt,
		PartialMetrics:    partialMetrics,
	}, nil
}

func (l *QueryOperateOrderFunnelLogic) queryOrderCreatedBuckets(
	scope pkgscope.GovernanceScope,
	in *omsclient.QueryOperateOrderFunnelReq,
	startTime,
	endTime time.Time,
	bucket string,
) ([]operateOrderBucketRow, error) {
	return l.queryOrderBuckets("create_time", scope, in, startTime, endTime, bucket)
}

func (l *QueryOperateOrderFunnelLogic) queryPaySuccessBuckets(
	scope pkgscope.GovernanceScope,
	in *omsclient.QueryOperateOrderFunnelReq,
	startTime,
	endTime time.Time,
	bucket string,
) ([]operateOrderBucketRow, error) {
	return l.queryOrderBuckets("pay_time", scope, in, startTime, endTime, bucket)
}

func (l *QueryOperateOrderFunnelLogic) queryOrderBuckets(
	timeColumn string,
	scope pkgscope.GovernanceScope,
	in *omsclient.QueryOperateOrderFunnelReq,
	startTime,
	endTime time.Time,
	bucket string,
) ([]operateOrderBucketRow, error) {
	rows := make([]operateOrderBucketRow, 0)
	query := l.svcCtx.DB.WithContext(l.ctx).
		Table("oms_order_main o").
		Select(operateOrderBucketExpr("o."+timeColumn, bucket)+" AS bucket_start, COUNT(*) AS "+operateOrderMetricAlias(timeColumn)).
		Where("o.is_deleted = 0").
		Where("o."+timeColumn+" IS NOT NULL").
		Where("o."+timeColumn+" >= ? AND o."+timeColumn+" < ?", startTime, endTime).
		Where("o.platform_id = ? AND o.tenant_id = ? AND o.merchant_id = ?", scope.PlatformID, scope.TenantID, scope.MerchantID)

	if timeColumn == "pay_time" {
		query = query.Where("o.order_status = 2")
	}

	if channel, ok := operatefunnel.OptionalChannel(in.Channel); ok {
		query = query.Where(operateOrderChannelCaseSQL()+" = ?", channel)
	}

	if activityType, ok := operatefunnel.OptionalActivityType(in.ActivityType); ok {
		if activityType == operatefunnel.ActivityNone {
			query = query.Where("COALESCE(NULLIF(o.activity_type, ''), 'none') = 'none'")
		} else {
			query = query.Where("o.activity_type = ?", activityType)
			if in.ActivityId > 0 {
				query = query.Where("o.activity_id = ?", in.ActivityId)
			}
		}
	}

	err := query.Group("bucket_start").Order("bucket_start ASC").Scan(&rows).Error
	return rows, err
}

func (l *QueryOperateOrderFunnelLogic) queryOrderTrackingStart(column string) (sql.NullTime, bool, error) {
	var value sql.NullTime
	err := l.svcCtx.DB.WithContext(l.ctx).
		Table("oms_order_main").
		Select("MIN(" + column + ")").
		Where("is_deleted = 0").
		Where(column + " IS NOT NULL").
		Where("activity_type <> '' AND activity_type <> 'none'").
		Scan(&value).Error
	return value, value.Valid, err
}

func operateOrderBucketExpr(column, bucket string) string {
	switch bucket {
	case operatefunnel.BucketHour:
		return "DATE_FORMAT(" + column + ", '%Y-%m-%d %H:00:00')"
	default:
		return "DATE_FORMAT(" + column + ", '%Y-%m-%d 00:00:00')"
	}
}

func operateOrderMetricAlias(column string) string {
	if column == "pay_time" {
		return "pay_success"
	}
	return "order_created"
}

func operateOrderChannelCaseSQL() string {
	return "CASE o.source_type WHEN 1 THEN 'app' WHEN 2 THEN 'pc' WHEN 3 THEN 'mini_program' ELSE 'unknown' END"
}
