package draw_activity

import (
	"context"

	"github.com/feihua/zero-admin/api/admin/internal/common"
	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"
	"github.com/zeromicro/go-zero/core/logx"
)

type QueryDrawActivityListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryDrawActivityListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryDrawActivityListLogic {
	return &QueryDrawActivityListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QueryDrawActivityListLogic) QueryDrawActivityList(req *types.QueryDrawActivityListReq) (*types.QueryDrawActivityListResp, error) {
	queryScope, err := common.ResolveQueryGovernanceScope(l.ctx, common.RequestedGovernanceScope{
		ScopeType:  req.ScopeType,
		PlatformID: req.PlatformId,
		TenantID:   req.TenantId,
		MerchantID: req.MerchantId,
	})
	if err != nil {
		return nil, err
	}
	result, err := l.svcCtx.DrawActivityService.QueryDrawActivityList(l.ctx, &smsclient.QueryDrawActivityListReq{
		PageNum:          req.Current,
		PageSize:         req.PageSize,
		Name:             req.Name,
		Status:           req.Status,
		AuditStatus:      req.AuditStatus,
		PublishReadiness: req.PublishReadiness,
		Scope:            common.SMSGovernanceScope(queryScope),
	})
	if err != nil {
		return nil, grpcDrawError(err)
	}
	list := make([]types.QueryDrawActivityListData, 0, len(result.List))
	for _, item := range result.List {
		if item == nil {
			continue
		}
		list = append(list, types.QueryDrawActivityListData{
			Id:                    item.Id,
			ActivityCode:          item.ActivityCode,
			Name:                  item.Name,
			StartTime:             item.StartTime,
			EndTime:               item.EndTime,
			Status:                item.Status,
			AuditStatus:           item.AuditStatus,
			IsEnabled:             item.IsEnabled,
			PublishReadiness:      item.PublishReadiness,
			PublishFailureSummary: item.PublishFailureSummary,
			ScopeType:             item.ScopeType,
			PlatformId:            item.PlatformId,
			TenantId:              item.TenantId,
			MerchantId:            item.MerchantId,
			UpdateTime:            item.UpdateTime,
			ShowOnHome:            item.ShowOnHome,
			HomeEntryTitle:        item.HomeEntryTitle,
			ReadinessLabel:        item.ReadinessLabel,
		})
	}
	return &types.QueryDrawActivityListResp{
		Code:     "000000",
		Message:  "查询抽卡活动列表成功",
		Current:  req.Current,
		Data:     list,
		PageSize: req.PageSize,
		Success:  true,
		Total:    result.Total,
	}, nil
}
