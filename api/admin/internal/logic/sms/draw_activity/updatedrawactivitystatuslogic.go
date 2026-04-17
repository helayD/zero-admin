package draw_activity

import (
	"context"

	"github.com/feihua/zero-admin/api/admin/internal/common"
	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"
	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateDrawActivityStatusLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateDrawActivityStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateDrawActivityStatusLogic {
	return &UpdateDrawActivityStatusLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateDrawActivityStatusLogic) UpdateDrawActivityStatus(req *types.UpdateDrawActivityStatusReq) (*types.BaseResp, error) {
	userId, err := common.GetUserId(l.ctx)
	if err != nil {
		return nil, err
	}
	userName, err := common.GetUserName(l.ctx)
	if err != nil {
		return nil, err
	}
	writeScope, err := common.ResolveWriteGovernanceScope(l.ctx, common.RequestedGovernanceScope{
		ScopeType:  req.ScopeType,
		PlatformID: req.PlatformId,
		TenantID:   req.TenantId,
		MerchantID: req.MerchantId,
	})
	if err != nil {
		return nil, err
	}
	_, err = l.svcCtx.DrawActivityService.UpdateDrawActivityStatus(l.ctx, &smsclient.UpdateDrawActivityStatusReq{
		Ids:          req.Ids,
		Status:       req.Status,
		Scope:        common.SMSGovernanceScope(writeScope),
		UpdateBy:     userId,
		OperatorName: userName,
	})
	if err != nil {
		return nil, grpcDrawError(err)
	}
	return &types.BaseResp{Code: "000000", Message: "更新抽卡活动状态成功"}, nil
}
