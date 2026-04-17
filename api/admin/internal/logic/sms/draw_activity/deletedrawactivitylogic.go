package draw_activity

import (
	"context"

	"github.com/feihua/zero-admin/api/admin/internal/common"
	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"
	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteDrawActivityLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteDrawActivityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteDrawActivityLogic {
	return &DeleteDrawActivityLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteDrawActivityLogic) DeleteDrawActivity(req *types.DeleteDrawActivityReq) (*types.BaseResp, error) {
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
	_, err = l.svcCtx.DrawActivityService.DeleteDrawActivity(l.ctx, &smsclient.DeleteDrawActivityReq{
		Ids:          req.Ids,
		Scope:        common.SMSGovernanceScope(writeScope),
		UpdateBy:     userId,
		OperatorName: userName,
	})
	if err != nil {
		return nil, grpcDrawError(err)
	}
	return &types.BaseResp{Code: "000000", Message: "删除抽卡活动成功"}, nil
}
