package order_main

import (
	"context"
	"github.com/feihua/zero-admin/api/admin/internal/common"
	"github.com/feihua/zero-admin/api/admin/internal/common/errorx"
	"github.com/feihua/zero-admin/api/admin/internal/common/res"
	"github.com/feihua/zero-admin/rpc/oms/omsclient"
	"github.com/zeromicro/go-zero/core/logc"

	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CloseOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCloseOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CloseOrderLogic {
	return &CloseOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CloseOrderLogic) CloseOrder(req *types.CloseOrderReq) (resp *types.BaseResp, err error) {
	userId, err := common.GetUserId(l.ctx)
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
	_, err = l.svcCtx.OrderService.CloseOrder(l.ctx, &omsclient.CloseOrderReq{
		Ids:        req.Ids,
		Note:       req.Note,
		OperatorId: userId,
		Scope:      common.OMSGovernanceScope(writeScope),
	})

	if err != nil {
		logc.Errorf(l.ctx, "批量关闭订单失败,参数：%+v,响应：%s", req, err.Error())
		return nil, errorx.NewDefaultError("批量关闭订单失败")
	}

	return res.Success()
}
