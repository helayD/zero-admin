package order_main

import (
	"context"

	"github.com/feihua/zero-admin/api/admin/internal/common"
	"github.com/feihua/zero-admin/api/admin/internal/common/errorx"
	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/feihua/zero-admin/rpc/oms/omsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"google.golang.org/grpc/status"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateMoneyInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateMoneyInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateMoneyInfoLogic {
	return &UpdateMoneyInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateMoneyInfoLogic) UpdateMoneyInfo(req *types.UpdateMoneyInfoReq) (resp *types.BaseResp, err error) {
	writeScope, err := common.ResolveWriteGovernanceScope(l.ctx, common.RequestedGovernanceScope{
		ScopeType:  req.ScopeType,
		PlatformID: req.PlatformId,
		TenantID:   req.TenantId,
		MerchantID: req.MerchantId,
	})
	if err != nil {
		return nil, err
	}
	_, err = l.svcCtx.OrderService.UpdateOrder(l.ctx, &omsclient.UpdateOrderReq{
		Id:             req.Id,
		OrderStatus:    req.Status,
		FreightAmount:  float32(req.FreightAmount),
		DiscountAmount: float32(req.DiscountAmount),
		Scope:          common.OMSGovernanceScope(writeScope),
	})
	if err != nil {
		logc.Errorf(l.ctx, "更新订单费用失败,参数：%+v,响应：%s", req, err.Error())
		s, _ := status.FromError(err)
		return nil, errorx.NewDefaultError(s.Message())
	}

	return &types.BaseResp{
		Code:    "000000",
		Message: "更新订单费用成功",
	}, nil
}
