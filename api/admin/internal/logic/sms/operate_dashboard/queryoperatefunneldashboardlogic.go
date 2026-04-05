package operate_dashboard

import (
	"context"
	"strings"
	"time"

	admincommon "github.com/feihua/zero-admin/api/admin/internal/common"
	"github.com/feihua/zero-admin/api/admin/internal/common/errorx"
	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/feihua/zero-admin/pkg/operatefunnel"
	"github.com/feihua/zero-admin/rpc/oms/omsclient"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/status"
)

type QueryOperateFunnelDashboardLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryOperateFunnelDashboardLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryOperateFunnelDashboardLogic {
	return &QueryOperateFunnelDashboardLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QueryOperateFunnelDashboardLogic) QueryOperateFunnelDashboard(req *types.QueryOperateFunnelDashboardReq) (resp *types.QueryOperateFunnelDashboardResp, err error) {
	queryScope, err := admincommon.ResolveQueryGovernanceScope(l.ctx, admincommon.RequestedGovernanceScope{
		ScopeType:  req.ScopeType,
		PlatformID: req.PlatformId,
		TenantID:   req.TenantId,
		MerchantID: req.MerchantId,
	})
	if err != nil {
		return nil, errorx.NewDefaultError(err.Error())
	}

	startText, endText := normalizeDashboardTimeRange(req.StartTime, req.EndTime)
	startTime, endTime, err := operatefunnel.ParseTimeRange(startText, endText)
	if err != nil {
		return nil, errorx.NewDefaultError(err.Error())
	}
	bucket := operatefunnel.NormalizeBucket(req.Bucket, startTime, endTime)

	trafficResp, err := l.svcCtx.OperateDashboardService.QueryOperateTrafficFunnel(l.ctx, &smsclient.QueryOperateTrafficFunnelReq{
		Scope:        admincommon.SMSGovernanceScope(queryScope),
		StartTime:    startText,
		EndTime:      endText,
		Channel:      req.Channel,
		ActivityType: req.ActivityType,
		ActivityId:   req.ActivityId,
		Bucket:       bucket,
	})
	if err != nil {
		logc.Errorf(l.ctx, "查询经营漏斗流量聚合失败, req=%+v, err=%s", req, err.Error())
		return nil, errorx.NewDefaultError(rpcErrorMessage(err))
	}

	cartResp, err := l.svcCtx.CartItemService.QueryOperateCartFunnel(l.ctx, &omsclient.QueryOperateCartFunnelReq{
		Scope:        admincommon.OMSGovernanceScope(queryScope),
		StartTime:    startText,
		EndTime:      endText,
		Channel:      req.Channel,
		ActivityType: req.ActivityType,
		ActivityId:   req.ActivityId,
		Bucket:       bucket,
	})
	if err != nil {
		logc.Errorf(l.ctx, "查询经营漏斗加购聚合失败, req=%+v, err=%s", req, err.Error())
		return nil, errorx.NewDefaultError(rpcErrorMessage(err))
	}

	orderResp, err := l.svcCtx.OrderService.QueryOperateOrderFunnel(l.ctx, &omsclient.QueryOperateOrderFunnelReq{
		Scope:        admincommon.OMSGovernanceScope(queryScope),
		StartTime:    startText,
		EndTime:      endText,
		Channel:      req.Channel,
		ActivityType: req.ActivityType,
		ActivityId:   req.ActivityId,
		Bucket:       bucket,
	})
	if err != nil {
		logc.Errorf(l.ctx, "查询经营漏斗订单聚合失败, req=%+v, err=%s", req, err.Error())
		return nil, errorx.NewDefaultError(rpcErrorMessage(err))
	}

	couponResp, err := l.svcCtx.OperateDashboardService.QueryOperateCouponRedeem(l.ctx, &smsclient.QueryOperateCouponRedeemReq{
		Scope:        admincommon.SMSGovernanceScope(queryScope),
		StartTime:    startText,
		EndTime:      endText,
		Channel:      req.Channel,
		ActivityType: req.ActivityType,
		ActivityId:   req.ActivityId,
		Bucket:       bucket,
	})
	if err != nil {
		logc.Errorf(l.ctx, "查询经营漏斗核销聚合失败, req=%+v, err=%s", req, err.Error())
		return nil, errorx.NewDefaultError(rpcErrorMessage(err))
	}

	activityOptionsResp, err := l.svcCtx.OperateDashboardService.QueryOperateActivityOptions(l.ctx, &smsclient.QueryOperateActivityOptionsReq{
		Scope: admincommon.SMSGovernanceScope(queryScope),
	})
	if err != nil {
		logc.Errorf(l.ctx, "查询经营漏斗活动选项失败, req=%+v, err=%s", req, err.Error())
		return nil, errorx.NewDefaultError(rpcErrorMessage(err))
	}

	overview := types.OperateFunnelOverview{
		Exposure:         trafficResp.TotalExposure,
		Click:            trafficResp.TotalClick,
		AddCart:          cartResp.TotalAddCart,
		OrderCreated:     orderResp.TotalOrderCreated,
		PaySuccess:       orderResp.TotalPaySuccess,
		CouponRedeem:     couponResp.TotalCouponRedeem,
		ClickRate:        operatefunnel.SafeRate(trafficResp.TotalClick, trafficResp.TotalExposure),
		AddCartRate:      operatefunnel.SafeRate(cartResp.TotalAddCart, trafficResp.TotalClick),
		OrderRate:        operatefunnel.SafeRate(orderResp.TotalOrderCreated, cartResp.TotalAddCart),
		PayRate:          operatefunnel.SafeRate(orderResp.TotalPaySuccess, orderResp.TotalOrderCreated),
		CouponRedeemRate: operatefunnel.SafeRate(couponResp.TotalCouponRedeem, orderResp.TotalPaySuccess),
	}
	overview.Cards = buildOverviewCards(overview)

	series := buildOperateSeries(startTime, endTime, bucket, trafficResp.Buckets, cartResp.Buckets, orderResp.Buckets, couponResp.Buckets)
	activityOptions := buildActivityOptions(activityOptionsResp.List)

	return &types.QueryOperateFunnelDashboardResp{
		Code:    "000000",
		Message: "查询经营漏斗看板成功",
		Success: true,
		Data: types.QueryOperateFunnelDashboardData{
			Overview:        overview,
			Series:          series,
			ActivityOptions: activityOptions,
			TrackingStartedAt: latestTrackingStartedAt(
				trafficResp.TrackingStartedAt,
				cartResp.TrackingStartedAt,
				orderResp.TrackingStartedAt,
				couponResp.TrackingStartedAt,
			),
			PartialMetrics: mergePartialMetrics(
				trafficResp.PartialMetrics,
				cartResp.PartialMetrics,
				orderResp.PartialMetrics,
				couponResp.PartialMetrics,
			),
			Bucket: bucket,
		},
	}, nil
}

func normalizeDashboardTimeRange(startText, endText string) (string, string) {
	now := time.Now().In(operatefunnel.Location())
	defaultStart := operatefunnel.BucketStart(now, operatefunnel.BucketDay).AddDate(0, 0, -6)
	defaultEnd := operatefunnel.NextBucketStart(now, operatefunnel.BucketDay)

	start := strings.TrimSpace(startText)
	if start == "" {
		start = operatefunnel.FormatDateTime(defaultStart)
	}

	end := strings.TrimSpace(endText)
	if end == "" {
		end = operatefunnel.FormatDateTime(defaultEnd)
	}

	return start, end
}

func buildOverviewCards(overview types.OperateFunnelOverview) []types.OperateFunnelMetricCard {
	return []types.OperateFunnelMetricCard{
		{Key: operatefunnel.EventExposure, Label: "曝光", Value: overview.Exposure},
		{Key: operatefunnel.EventClick, Label: "点击", Value: overview.Click, Rate: overview.ClickRate, RateLabel: "点击率"},
		{Key: operatefunnel.EventAddCart, Label: "加购", Value: overview.AddCart, Rate: overview.AddCartRate, RateLabel: "加购率"},
		{Key: operatefunnel.EventOrderCreated, Label: "下单", Value: overview.OrderCreated, Rate: overview.OrderRate, RateLabel: "下单率"},
		{Key: operatefunnel.EventPaySuccess, Label: "支付", Value: overview.PaySuccess, Rate: overview.PayRate, RateLabel: "支付率"},
		{Key: operatefunnel.EventCouponRedeem, Label: "核销", Value: overview.CouponRedeem, Rate: overview.CouponRedeemRate, RateLabel: "核销率"},
	}
}

func buildOperateSeries(
	startTime, endTime time.Time,
	bucket string,
	trafficBuckets []*smsclient.OperateTrafficBucketPoint,
	cartBuckets []*omsclient.OperateCartBucketPoint,
	orderBuckets []*omsclient.OperateOrderBucketPoint,
	couponBuckets []*smsclient.OperateCouponRedeemBucketPoint,
) []types.OperateFunnelSeriesPoint {
	trafficByBucket := make(map[string]*smsclient.OperateTrafficBucketPoint, len(trafficBuckets))
	for _, bucketPoint := range trafficBuckets {
		trafficByBucket[bucketPoint.BucketStart] = bucketPoint
	}

	cartByBucket := make(map[string]*omsclient.OperateCartBucketPoint, len(cartBuckets))
	for _, bucketPoint := range cartBuckets {
		cartByBucket[bucketPoint.BucketStart] = bucketPoint
	}

	orderByBucket := make(map[string]*omsclient.OperateOrderBucketPoint, len(orderBuckets))
	for _, bucketPoint := range orderBuckets {
		orderByBucket[bucketPoint.BucketStart] = bucketPoint
	}

	couponByBucket := make(map[string]*smsclient.OperateCouponRedeemBucketPoint, len(couponBuckets))
	for _, bucketPoint := range couponBuckets {
		couponByBucket[bucketPoint.BucketStart] = bucketPoint
	}

	windows := operatefunnel.BuildBucketWindows(startTime, endTime, bucket)
	series := make([]types.OperateFunnelSeriesPoint, 0, len(windows))
	for _, window := range windows {
		key := operatefunnel.FormatDateTime(window.Start)
		trafficBucket := trafficByBucket[key]
		cartBucket := cartByBucket[key]
		orderBucket := orderByBucket[key]
		couponBucket := couponByBucket[key]

		exposure := int64(0)
		click := int64(0)
		addCart := int64(0)
		orderCreated := int64(0)
		paySuccess := int64(0)
		couponRedeem := int64(0)

		if trafficBucket != nil {
			exposure = trafficBucket.Exposure
			click = trafficBucket.Click
		}
		if cartBucket != nil {
			addCart = cartBucket.AddCart
		}
		if orderBucket != nil {
			orderCreated = orderBucket.OrderCreated
			paySuccess = orderBucket.PaySuccess
		}
		if couponBucket != nil {
			couponRedeem = couponBucket.CouponRedeem
		}

		series = append(series, types.OperateFunnelSeriesPoint{
			BucketLabel:      window.Label,
			BucketStart:      key,
			BucketEnd:        operatefunnel.FormatDateTime(window.End),
			Exposure:         exposure,
			Click:            click,
			AddCart:          addCart,
			OrderCreated:     orderCreated,
			PaySuccess:       paySuccess,
			CouponRedeem:     couponRedeem,
			ClickRate:        operatefunnel.SafeRate(click, exposure),
			AddCartRate:      operatefunnel.SafeRate(addCart, click),
			OrderRate:        operatefunnel.SafeRate(orderCreated, addCart),
			PayRate:          operatefunnel.SafeRate(paySuccess, orderCreated),
			CouponRedeemRate: operatefunnel.SafeRate(couponRedeem, paySuccess),
		})
	}

	return series
}

func buildActivityOptions(list []*smsclient.OperateFunnelActivityOption) []types.OperateFunnelActivityOption {
	options := make([]types.OperateFunnelActivityOption, 0, len(list))
	for _, item := range list {
		options = append(options, types.OperateFunnelActivityOption{
			ActivityType:  item.ActivityType,
			ActivityId:    item.ActivityId,
			ActivityName:  item.ActivityName,
			ActivityLabel: item.ActivityLabel,
		})
	}

	return options
}

func latestTrackingStartedAt(values ...string) string {
	var latest time.Time
	for _, value := range values {
		if strings.TrimSpace(value) == "" {
			continue
		}

		parsed, err := operatefunnel.ParseTime(value)
		if err != nil {
			continue
		}
		if latest.IsZero() || latest.Before(parsed) {
			latest = parsed
		}
	}

	if latest.IsZero() {
		return ""
	}

	return operatefunnel.FormatDateTime(latest)
}

func mergePartialMetrics(metricGroups ...[]string) []string {
	seen := make(map[string]struct{})
	for _, metrics := range metricGroups {
		for _, metric := range metrics {
			seen[metric] = struct{}{}
		}
	}

	merged := make([]string, 0, len(seen))
	for _, metric := range operatefunnel.FunnelStages() {
		if _, ok := seen[metric]; ok {
			merged = append(merged, metric)
		}
	}

	return merged
}

func rpcErrorMessage(err error) string {
	s, _ := status.FromError(err)
	if strings.TrimSpace(s.Message()) != "" {
		return s.Message()
	}
	return err.Error()
}
