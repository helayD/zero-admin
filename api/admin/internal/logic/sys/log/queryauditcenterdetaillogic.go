package log

import (
	"context"

	admincommon "github.com/feihua/zero-admin/api/admin/internal/common"
	"github.com/feihua/zero-admin/api/admin/internal/common/errorx"
	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/feihua/zero-admin/rpc/sys/sysclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/status"
)

type QueryAuditCenterDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryAuditCenterDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryAuditCenterDetailLogic {
	return &QueryAuditCenterDetailLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *QueryAuditCenterDetailLogic) QueryAuditCenterDetail(req *types.QueryAuditCenterDetailReq) (resp *types.QueryAuditCenterDetailResp, err error) {
	queryScope, err := admincommon.ResolveQueryGovernanceScope(l.ctx, admincommon.RequestedGovernanceScope{ScopeType: req.ScopeType, PlatformID: req.PlatformId, TenantID: req.TenantId, MerchantID: req.MerchantId})
	if err != nil {
		return nil, errorx.NewDefaultError(err.Error())
	}
	result, err := l.svcCtx.AuditCenterService.QueryAuditCenterDetail(l.ctx, &sysclient.QueryAuditCenterDetailReq{Scope: &sysclient.GovernanceScope{ScopeType: queryScope.ScopeType, PlatformId: queryScope.PlatformID, TenantId: queryScope.TenantID, MerchantId: queryScope.MerchantID, ScopeLabel: queryScope.Label()}, SourceType: req.SourceType, SourceId: req.SourceId})
	if err != nil {
		logc.Errorf(l.ctx, "查询治理审计详情失败, 参数:%+v, 异常:%s", req, err.Error())
		s, _ := status.FromError(err)
		return nil, errorx.NewDefaultError(s.Message())
	}
	timeline := make([]*types.AuditTimelineItem, 0, len(result.Timeline))
	for _, item := range result.Timeline {
		timeline = append(timeline, &types.AuditTimelineItem{SourceType: item.SourceType, SourceId: item.SourceId, TraceId: item.TraceId, EventType: item.EventType, Action: item.Action, Result: item.Result, OperatorName: item.OperatorName, ResourceType: item.ResourceType, ResourceId: item.ResourceId, ResourceName: item.ResourceName, RequestSummary: item.RequestSummary, HappenedAt: item.HappenedAt})
	}
	return &types.QueryAuditCenterDetailResp{Code: "000000", Message: "查询治理审计详情成功", Data: types.QueryAuditCenterDetailData{SourceType: result.SourceType, SourceId: result.SourceId, TraceId: result.TraceId, EventType: result.EventType, Action: result.Action, Result: result.Result, ScopeType: result.ScopeType, ScopeLabel: result.ScopeLabel, PlatformId: result.PlatformId, TenantId: result.TenantId, MerchantId: result.MerchantId, OperatorId: result.OperatorId, OperatorName: result.OperatorName, ResourceType: result.ResourceType, ResourceId: result.ResourceId, ResourceName: result.ResourceName, SubjectInfo: result.SubjectInfo, RequestSummary: result.RequestSummary, DetailPayload: result.DetailPayload, SensitiveMasked: result.SensitiveMasked, HappenedAt: result.HappenedAt, Timeline: timeline}}, nil
}
