package operate_dashboard

import (
	"context"

	admincommon "github.com/feihua/zero-admin/api/admin/internal/common"
	"github.com/feihua/zero-admin/api/admin/internal/common/errorx"
	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/feihua/zero-admin/pkg/operatefunnel"
	"github.com/feihua/zero-admin/rpc/oms/omsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
)

type QueryRepeatPurchaseAnalysisLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryRepeatPurchaseAnalysisLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryRepeatPurchaseAnalysisLogic {
	return &QueryRepeatPurchaseAnalysisLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QueryRepeatPurchaseAnalysisLogic) QueryRepeatPurchaseAnalysis(req *types.QueryRepeatPurchaseAnalysisReq) (*types.QueryRepeatPurchaseAnalysisResp, error) {
	queryScope, err := admincommon.ResolveQueryGovernanceScope(l.ctx, admincommon.RequestedGovernanceScope{
		ScopeType:  req.ScopeType,
		PlatformID: req.PlatformId,
		TenantID:   req.TenantId,
		MerchantID: req.MerchantId,
	})
	if err != nil {
		return nil, errorx.NewDefaultError(err.Error())
	}
	startText, endText, startTime, endTime, err := normalizeRepeatPurchaseTimeRange(req.StartTime, req.EndTime)
	if err != nil {
		return nil, errorx.NewDefaultError(err.Error())
	}
	bucket := operatefunnel.NormalizeBucket(req.Bucket, startTime, endTime)
	pageNum, pageSize := normalizeRepeatPurchasePage(req.PageNum, req.PageSize)

	analysisResp, err := l.svcCtx.OrderService.QueryRepeatPurchaseAnalysis(l.ctx, &omsclient.QueryRepeatPurchaseAnalysisReq{
		Scope:        admincommon.OMSGovernanceScope(queryScope),
		StartTime:    startText,
		EndTime:      endText,
		Channel:      req.Channel,
		ActivityType: req.ActivityType,
		ActivityId:   req.ActivityId,
		Bucket:       bucket,
	})
	if err != nil {
		logc.Errorf(l.ctx, "查询复购分析总览失败, req=%+v, err=%s", req, err.Error())
		return nil, errorx.NewDefaultError(rpcErrorMessage(err))
	}
	detailResp, err := l.svcCtx.OrderService.QueryRepeatPurchaseDetailList(l.ctx, &omsclient.QueryRepeatPurchaseDetailListReq{
		Scope:        admincommon.OMSGovernanceScope(queryScope),
		StartTime:    startText,
		EndTime:      endText,
		Channel:      req.Channel,
		ActivityType: req.ActivityType,
		ActivityId:   req.ActivityId,
		PageNum:      pageNum,
		PageSize:     pageSize,
	})
	if err != nil {
		logc.Errorf(l.ctx, "查询复购分析详情失败, req=%+v, err=%s", req, err.Error())
		return nil, errorx.NewDefaultError(rpcErrorMessage(err))
	}
	activityOptions, err := queryRepeatPurchaseActivityOptions(l.ctx, l.svcCtx.OperateDashboardService, admincommon.SMSGovernanceScope(queryScope))
	if err != nil {
		logc.Errorf(l.ctx, "查询复购分析活动选项失败, req=%+v, err=%s", req, err.Error())
		return nil, errorx.NewDefaultError(rpcErrorMessage(err))
	}
	briefMap, err := queryRepeatPurchaseMemberBriefMap(l.ctx, l.svcCtx.MemberInfoService, detailResp.List)
	if err != nil {
		logc.Errorf(l.ctx, "查询复购分析会员简要信息失败, req=%+v, err=%s", req, err.Error())
		return nil, errorx.NewDefaultError(rpcErrorMessage(err))
	}

	return &types.QueryRepeatPurchaseAnalysisResp{
		Code:    "000000",
		Message: "查询复购分析成功",
		Success: true,
		Data: types.QueryRepeatPurchaseAnalysisData{
			Overview:          buildRepeatPurchaseOverview(analysisResp.Overview),
			Trends:            buildRepeatPurchaseTrendPoints(analysisResp.Trends),
			Details:           buildRepeatPurchaseDetailItems(detailResp.List, briefMap),
			Total:             detailResp.Total,
			PageNum:           pageNum,
			PageSize:          pageSize,
			ActivityOptions:   activityOptions,
			TrackingStartedAt: latestTrackingStartedAt(analysisResp.TrackingStartedAt, detailResp.TrackingStartedAt),
			PartialMetrics:    mergeRepeatPurchasePartialMetrics(analysisResp.PartialMetrics, detailResp.PartialMetrics),
			Bucket:            bucket,
		},
	}, nil
}
