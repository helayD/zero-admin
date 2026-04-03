package cartitemservicelogic

import (
	"context"
	"database/sql"

	"github.com/feihua/zero-admin/pkg/operatefunnel"
	logiccommon "github.com/feihua/zero-admin/rpc/oms/internal/logic/common"
	"github.com/feihua/zero-admin/rpc/oms/internal/svc"
	"github.com/feihua/zero-admin/rpc/oms/omsclient"
	"github.com/zeromicro/go-zero/core/logx"
)

type QueryOperateCartFunnelLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

type operateCartBucketRow struct {
	BucketStart string `gorm:"column:bucket_start"`
	AddCart     int64  `gorm:"column:add_cart"`
}

func NewQueryOperateCartFunnelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryOperateCartFunnelLogic {
	return &QueryOperateCartFunnelLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查询经营漏斗加购聚合
func (l *QueryOperateCartFunnelLogic) QueryOperateCartFunnel(in *omsclient.QueryOperateCartFunnelReq) (*omsclient.QueryOperateCartFunnelResp, error) {
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

	rows := make([]operateCartBucketRow, 0)
	query := l.svcCtx.DB.WithContext(l.ctx).
		Table("oms_cart_item c").
		Select(operateCartBucketExpr("c.create_time", bucket)+" AS bucket_start, COUNT(*) AS add_cart").
		Joins("JOIN pms_product_spu spu ON spu.id = c.product_id").
		Where("c.delete_status = 0").
		Where("c.create_time >= ? AND c.create_time < ?", startTime, endTime).
		Where("spu.platform_id = ? AND spu.tenant_id = ? AND spu.merchant_id = ?", scope.PlatformID, scope.TenantID, scope.MerchantID)

	if channel, ok := operatefunnel.OptionalChannel(in.Channel); ok {
		query = query.Where(operateCartChannelCaseSQL()+" = ?", channel)
	}

	activityRequested := false
	if activityType, ok := operatefunnel.OptionalActivityType(in.ActivityType); ok {
		activityRequested = activityType != operatefunnel.ActivityNone
		if activityType == operatefunnel.ActivityNone {
			query = query.Where("COALESCE(NULLIF(c.activity_type, ''), 'none') = 'none'")
		} else {
			query = query.Where("c.activity_type = ?", activityType)
			if in.ActivityId > 0 {
				query = query.Where("c.activity_id = ?", in.ActivityId)
			}
		}
	}

	if err = query.Group("bucket_start").Order("bucket_start ASC").Scan(&rows).Error; err != nil {
		return nil, err
	}

	countByBucket := make(map[string]int64, len(rows))
	for _, row := range rows {
		countByBucket[row.BucketStart] = row.AddCart
	}

	var (
		total          int64
		buckets        = make([]*omsclient.OperateCartBucketPoint, 0, len(windows))
		trackingAt     string
		partialMetrics []string
	)

	if activityRequested {
		trackingStart, ok, trackErr := l.queryCartActivityTrackingStart()
		if trackErr != nil {
			return nil, trackErr
		}
		if ok {
			trackingAt = operatefunnel.FormatDateTime(trackingStart.Time)
			if startTime.Before(trackingStart.Time) {
				partialMetrics = append(partialMetrics, operatefunnel.EventAddCart)
			}
		}
	}

	for _, window := range windows {
		key := operatefunnel.FormatDateTime(window.Start)
		count := countByBucket[key]
		total += count
		buckets = append(buckets, &omsclient.OperateCartBucketPoint{
			BucketLabel: window.Label,
			BucketStart: key,
			BucketEnd:   operatefunnel.FormatDateTime(window.End),
			AddCart:     count,
		})
	}

	return &omsclient.QueryOperateCartFunnelResp{
		Buckets:           buckets,
		TotalAddCart:      total,
		TrackingStartedAt: trackingAt,
		PartialMetrics:    partialMetrics,
	}, nil
}

func (l *QueryOperateCartFunnelLogic) queryCartActivityTrackingStart() (timeValue sql.NullTime, ok bool, err error) {
	err = l.svcCtx.DB.WithContext(l.ctx).
		Table("oms_cart_item").
		Select("MIN(create_time)").
		Where("delete_status = 0").
		Where("activity_type <> '' AND activity_type <> 'none'").
		Scan(&timeValue).Error
	return timeValue, timeValue.Valid, err
}

func operateCartBucketExpr(column, bucket string) string {
	switch bucket {
	case operatefunnel.BucketHour:
		return "DATE_FORMAT(" + column + ", '%Y-%m-%d %H:00:00')"
	default:
		return "DATE_FORMAT(" + column + ", '%Y-%m-%d 00:00:00')"
	}
}

func operateCartChannelCaseSQL() string {
	return "CASE c.source WHEN 1 THEN 'pc' WHEN 2 THEN 'h5' WHEN 3 THEN 'mini_program' WHEN 4 THEN 'app' ELSE 'unknown' END"
}
