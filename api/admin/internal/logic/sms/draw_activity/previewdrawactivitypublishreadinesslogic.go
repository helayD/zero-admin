package draw_activity

import (
	"context"

	"github.com/feihua/zero-admin/api/admin/internal/common"
	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"
	"github.com/zeromicro/go-zero/core/logx"
)

type PreviewDrawActivityPublishReadinessLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPreviewDrawActivityPublishReadinessLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PreviewDrawActivityPublishReadinessLogic {
	return &PreviewDrawActivityPublishReadinessLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PreviewDrawActivityPublishReadinessLogic) PreviewDrawActivityPublishReadiness(req *types.PreviewDrawActivityPublishReadinessReq) (*types.PreviewDrawActivityPublishReadinessResp, error) {
	queryScope, err := common.ResolveQueryGovernanceScope(l.ctx, common.RequestedGovernanceScope{
		ScopeType:  req.ScopeType,
		PlatformID: req.PlatformId,
		TenantID:   req.TenantId,
		MerchantID: req.MerchantId,
	})
	if err != nil {
		return nil, err
	}
	result, err := l.svcCtx.DrawActivityService.PreviewDrawActivityPublishReadiness(l.ctx, &smsclient.PreviewDrawActivityPublishReadinessReq{
		Id:    req.Id,
		Scope: common.SMSGovernanceScope(queryScope),
	})
	if err != nil {
		return nil, grpcDrawError(err)
	}
	return &types.PreviewDrawActivityPublishReadinessResp{
		Code:    "000000",
		Message: "抽卡活动发布预检成功",
		Data: types.PreviewDrawActivityPublishReadinessData{
			ReadyToPublish:   result.ReadyToPublish,
			PublishReadiness: result.PublishReadiness,
			ReadinessLabel:   result.ReadinessLabel,
			Summary:          result.Summary,
			Items:            toAPIReadinessItems(result.Items),
		},
	}, nil
}
