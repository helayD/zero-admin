package auditcenterservicelogic

import (
	"context"

	"github.com/feihua/zero-admin/rpc/sys/internal/svc"
	"github.com/feihua/zero-admin/rpc/sys/sysclient"
	"github.com/zeromicro/go-zero/core/logx"
)

type QueryAuditCenterListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewQueryAuditCenterListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryAuditCenterListLogic {
	return &QueryAuditCenterListLogic{ctx: ctx, svcCtx: svcCtx, Logger: logx.WithContext(ctx)}
}

func (l *QueryAuditCenterListLogic) QueryAuditCenterList(in *sysclient.QueryAuditCenterListReq) (*sysclient.QueryAuditCenterListResp, error) {
	items, err := queryAuditItems(l.ctx, l.svcCtx, in)
	if err != nil {
		return nil, err
	}
	pageNum := in.PageNum
	if pageNum <= 0 {
		pageNum = 1
	}
	pageSize := in.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	start := int((pageNum - 1) * pageSize)
	if start >= len(items) {
		return &sysclient.QueryAuditCenterListResp{Total: int64(len(items)), List: []*sysclient.AuditCenterListItem{}}, nil
	}
	end := start + int(pageSize)
	if end > len(items) {
		end = len(items)
	}
	list := make([]*sysclient.AuditCenterListItem, 0, end-start)
	for _, item := range items[start:end] {
		list = append(list, &sysclient.AuditCenterListItem{SourceType: item.SourceType, SourceId: item.SourceID, TraceId: item.TraceID, EventType: item.EventType, Action: item.Action, Result: item.Result, ScopeType: item.ScopeType, ScopeLabel: item.ScopeLabel, PlatformId: item.PlatformID, TenantId: item.TenantID, MerchantId: item.MerchantID, OperatorId: item.OperatorID, OperatorName: item.OperatorName, ResourceType: item.ResourceType, ResourceId: item.ResourceID, ResourceName: item.ResourceName, SubjectInfo: item.SubjectInfo, RequestSummary: item.RequestSummary, SensitiveMasked: item.SensitiveMasked, HappenedAt: item.HappenedAt.Format(contextTimeLayout())})
	}
	return &sysclient.QueryAuditCenterListResp{Total: int64(len(items)), List: list}, nil
}

func contextTimeLayout() string { return "2006-01-02 15:04:05" }
