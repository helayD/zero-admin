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

type QueryAuditCenterListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryAuditCenterListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryAuditCenterListLogic {
	return &QueryAuditCenterListLogic{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

func (l *QueryAuditCenterListLogic) QueryAuditCenterList(req *types.QueryAuditCenterListReq) (resp *types.QueryAuditCenterListResp, err error) {
	queryScope, err := admincommon.ResolveQueryGovernanceScope(l.ctx, admincommon.RequestedGovernanceScope{ScopeType: req.ScopeType, PlatformID: req.PlatformId, TenantID: req.TenantId, MerchantID: req.MerchantId})
	if err != nil {
		return nil, errorx.NewDefaultError(err.Error())
	}
	result, err := l.svcCtx.AuditCenterService.QueryAuditCenterList(l.ctx, &sysclient.QueryAuditCenterListReq{Scope: &sysclient.GovernanceScope{ScopeType: queryScope.ScopeType, PlatformId: queryScope.PlatformID, TenantId: queryScope.TenantID, MerchantId: queryScope.MerchantID, ScopeLabel: queryScope.Label()}, PageNum: req.Current, PageSize: req.PageSize, OperatorId: req.OperatorId, OperatorName: req.OperatorName, TraceId: req.TraceId, ResourceType: req.ResourceType, ResourceId: req.ResourceId, EventType: req.EventType, Result: req.Result, Keyword: req.Keyword, StartTime: req.StartTime, EndTime: req.EndTime})
	if err != nil {
		logc.Errorf(l.ctx, "查询治理审计中心列表失败, 参数:%+v, 异常:%s", req, err.Error())
		s, _ := status.FromError(err)
		return nil, errorx.NewDefaultError(s.Message())
	}
	list := make([]*types.QueryAuditCenterListItem, 0, len(result.List))
	for _, item := range result.List {
		list = append(list, &types.QueryAuditCenterListItem{SourceType: item.SourceType, SourceId: item.SourceId, TraceId: item.TraceId, EventType: item.EventType, Action: item.Action, Result: item.Result, ScopeType: item.ScopeType, ScopeLabel: item.ScopeLabel, PlatformId: item.PlatformId, TenantId: item.TenantId, MerchantId: item.MerchantId, OperatorId: item.OperatorId, OperatorName: item.OperatorName, ResourceType: item.ResourceType, ResourceId: item.ResourceId, ResourceName: item.ResourceName, SubjectInfo: item.SubjectInfo, RequestSummary: item.RequestSummary, SensitiveMasked: item.SensitiveMasked, HappenedAt: item.HappenedAt})
	}
	return &types.QueryAuditCenterListResp{Code: "000000", Message: "查询治理审计中心成功", Current: req.Current, Data: list, PageSize: req.PageSize, Success: true, Total: result.Total}, nil
}
