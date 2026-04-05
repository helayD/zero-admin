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

type QueryOperateTrafficFunnelLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewQueryOperateTrafficFunnelLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryOperateTrafficFunnelLogic {
	return &QueryOperateTrafficFunnelLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *QueryOperateTrafficFunnelLogic) QueryOperateTrafficFunnel(in *smsclient.QueryOperateTrafficFunnelReq) (*smsclient.QueryOperateTrafficFunnelResp, error) {
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

	rows := make([]trafficBucketRow, 0)
	query := l.buildOperateTrafficBucketQuery(scope, in, startTime, endTime, bucket)

	if err = query.Group("bucket_start").Order("bucket_start ASC").Scan(&rows).Error; err != nil {
		return nil, err
	}

	countByBucket := make(map[string]trafficBucketRow, len(rows))
	for _, row := range rows {
		countByBucket[row.BucketStart] = row
	}

	trackingStart, partialMetrics, err := l.queryTrafficTrackingState(scope, startTime, in)
	if err != nil {
		return nil, err
	}

	var (
		totalExposure int64
		totalClick    int64
		buckets       = make([]*smsclient.OperateTrafficBucketPoint, 0, len(windows))
	)
	for _, window := range windows {
		key := operatefunnel.FormatDateTime(window.Start)
		row := countByBucket[key]
		totalExposure += row.Exposure
		totalClick += row.Click
		buckets = append(buckets, &smsclient.OperateTrafficBucketPoint{
			BucketLabel: window.Label,
			BucketStart: key,
			BucketEnd:   operatefunnel.FormatDateTime(window.End),
			Exposure:    row.Exposure,
			Click:       row.Click,
		})
	}

	return &smsclient.QueryOperateTrafficFunnelResp{
		Buckets:           buckets,
		TotalExposure:     totalExposure,
		TotalClick:        totalClick,
		TrackingStartedAt: nullTimeString(trackingStart),
		PartialMetrics:    partialMetrics,
	}, nil
}

func (l *QueryOperateTrafficFunnelLogic) buildOperateTrafficBucketQuery(
	scope pkgscope.GovernanceScope,
	in *smsclient.QueryOperateTrafficFunnelReq,
	startTime, endTime time.Time,
	bucket string,
) *gorm.DB {
	scopeSQL, scopeArgs := scopeClause("e", scope)
	query := l.svcCtx.DB.WithContext(l.ctx).
		Table("sms_operate_funnel_event e").
		Select(operateBucketExpr("e.stat_time", bucket)+" AS bucket_start, "+
			"SUM(CASE WHEN e.event_type = 'exposure' THEN 1 ELSE 0 END) AS exposure, "+
			"SUM(CASE WHEN e.event_type = 'click' THEN 1 ELSE 0 END) AS click").
		Where("e.event_type IN ('exposure', 'click')").
		Where("e.stat_time >= ? AND e.stat_time < ?", startTime, endTime).
		Where(scopeSQL, scopeArgs...)

	if channel, ok := operatefunnel.OptionalChannel(in.Channel); ok {
		query = query.Where("e.channel = ?", channel)
	}
	if activityType, ok := operatefunnel.OptionalActivityType(in.ActivityType); ok {
		if activityType == operatefunnel.ActivityNone {
			query = query.Where("COALESCE(NULLIF(e.activity_type, ''), 'none') = 'none'")
		} else {
			query = query.Where("e.activity_type = ?", activityType)
			if in.ActivityId > 0 {
				query = query.Where("e.activity_id = ?", in.ActivityId)
			}
		}
	}

	return query
}

func (l *QueryOperateTrafficFunnelLogic) queryTrafficTrackingState(scope pkgscope.GovernanceScope, startTime time.Time, in *smsclient.QueryOperateTrafficFunnelReq) (sql.NullTime, []string, error) {
	var trackingStart sql.NullTime
	scopeSQL, scopeArgs := scopeClause("e", scope)
	query := l.svcCtx.DB.WithContext(l.ctx).
		Table("sms_operate_funnel_event e").
		Select("MIN(e.stat_time)").
		Where("e.event_type IN ('exposure', 'click')").
		Where(scopeSQL, scopeArgs...)

	if channel, ok := operatefunnel.OptionalChannel(in.Channel); ok {
		query = query.Where("e.channel = ?", channel)
	}
	if activityType, ok := operatefunnel.OptionalActivityType(in.ActivityType); ok {
		if activityType == operatefunnel.ActivityNone {
			query = query.Where("COALESCE(NULLIF(e.activity_type, ''), 'none') = 'none'")
		} else {
			query = query.Where("e.activity_type = ?", activityType)
			if in.ActivityId > 0 {
				query = query.Where("e.activity_id = ?", in.ActivityId)
			}
		}
	}

	if err := query.Scan(&trackingStart).Error; err != nil {
		return sql.NullTime{}, nil, err
	}

	return trackingStart, buildTrafficPartialMetrics(trackingStart, startTime), nil
}

func buildTrafficPartialMetrics(trackingStart sql.NullTime, startTime time.Time) []string {
	partialMetrics := make([]string, 0, 2)
	if trackingStart.Valid && startTime.Before(trackingStart.Time) {
		partialMetrics = append(partialMetrics, operatefunnel.EventExposure, operatefunnel.EventClick)
	}
	return partialMetrics
}
