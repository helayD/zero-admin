package order

import (
	"context"

	"github.com/feihua/zero-admin/api/front/internal/logic/common"
	"github.com/feihua/zero-admin/api/front/internal/svc"
	"github.com/feihua/zero-admin/api/front/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

// UpdateOrderStatusLogic 统一更新订单状态
// 调用 OrderStatusService 执行：归属校验 → 幂等校验 → 状态转换校验 → OMS RPC 更新 → 支付状态更新 → 操作日志
//
// Author: Claude AI
// Date: 2026-03-30
// Story: 6-5 Task 8
type UpdateOrderStatusLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateOrderStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateOrderStatusLogic {
	return &UpdateOrderStatusLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// UpdateOrderStatus 统一更新订单状态入口
func (l *UpdateOrderStatusLogic) UpdateOrderStatus(req *types.UpdateOrderStatusReq) (resp *types.UpdateOrderStatusResp, err error) {
	memberId, err := common.GetMemberId(l.ctx)
	if err != nil {
		return nil, err
	}

	orderSvc := NewOrderStatusService(l.ctx, l.svcCtx)

	// 构造内部请求
	internalReq := &UpdateOrderStatusReq{
		OrderId:  req.OrderId,
		MemberId: memberId,
		Action:   req.Action,
		BizData:  req.BizData,
	}

	result, err := orderSvc.UpdateOrderStatus(internalReq)
	if err != nil {
		return nil, err
	}

	return &types.UpdateOrderStatusResp{
		Code:      result.Code,
		Message:   result.Message,
		OldStatus: result.OldStatus,
		NewStatus: result.NewStatus,
		PayStatus: result.PayStatus,
		UpdatedAt: result.UpdatedAt,
	}, nil
}
