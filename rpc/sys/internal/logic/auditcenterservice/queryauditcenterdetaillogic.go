package auditcenterservicelogic

import (
	"context"
	"time"

	"github.com/feihua/zero-admin/rpc/sys/internal/svc"
	"github.com/feihua/zero-admin/rpc/sys/sysclient"
	"github.com/zeromicro/go-zero/core/logx"
)

type QueryAuditCenterDetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewQueryAuditCenterDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryAuditCenterDetailLogic {
	return &QueryAuditCenterDetailLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *QueryAuditCenterDetailLogic) QueryAuditCenterDetail(in *sysclient.QueryAuditCenterDetailReq) (*sysclient.QueryAuditCenterDetailResp, error) {
	items, err := queryAuditItems(l.ctx, l.svcCtx, &sysclient.QueryAuditCenterListReq{Scope: in.Scope, PageNum: 1, PageSize: 1000, StartTime: time.Now().Add(-maxAuditWindow).Format(contextTimeLayout()), EndTime: time.Now().Format(contextTimeLayout())})
	if err != nil {
		return nil, err
	}
	current, err := findAuditItem(items, in.SourceType, in.SourceId)
	if err != nil {
		return nil, err
	}
	return &sysclient.QueryAuditCenterDetailResp{SourceType: current.SourceType, SourceId: current.SourceID, TraceId: current.TraceID, EventType: current.EventType, Action: current.Action, Result: current.Result, ScopeType: current.ScopeType, ScopeLabel: current.ScopeLabel, PlatformId: current.PlatformID, TenantId: current.TenantID, MerchantId: current.MerchantID, OperatorId: current.OperatorID, OperatorName: current.OperatorName, ResourceType: current.ResourceType, ResourceId: current.ResourceID, ResourceName: current.ResourceName, SubjectInfo: current.SubjectInfo, RequestSummary: current.RequestSummary, DetailPayload: current.DetailPayload, SensitiveMasked: current.SensitiveMasked, HappenedAt: current.HappenedAt.Format(contextTimeLayout()), Timeline: buildTimeline(items, current)}, nil
}
