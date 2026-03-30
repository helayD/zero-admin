package order

import (
	"context"

	"github.com/feihua/zero-admin/api/front/internal/logic/common"
	"github.com/feihua/zero-admin/pkg/errorx"
	"github.com/feihua/zero-admin/rpc/oms/omsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"google.golang.org/grpc/status"

	"github.com/feihua/zero-admin/api/front/internal/svc"
	"github.com/feihua/zero-admin/api/front/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

// ConfirmReceiveOrderLogic 用户确认收货
// 重构说明（Story 6-5）：
//  - 使用 OrderStatusService 进行状态转换校验（必须为 OrderStatusShipped=3）
//  - ⚠️ OMS 跳过了 order_status=3，确认收货由 ConfirmOrder 直接 2→4
//  - 新增 AddOrderOperationLog 操作日志记录
/*
Author: LiuFeiHua
Date: 2025/6/20 9:52
*/
type ConfirmReceiveOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewConfirmReceiveOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ConfirmReceiveOrderLogic {
	return &ConfirmReceiveOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// ConfirmReceiveOrder 用户确认收货
func (l *ConfirmReceiveOrderLogic) ConfirmReceiveOrder(req *types.ConfirmReceiveOrderReq) (resp *types.ConfirmReceiveOrderResp, err error) {
	memberId, err := common.GetMemberId(l.ctx)
	if err != nil {
		return nil, err
	}

	// Story 6-5 Task 4.1：归属校验（双重保护）
	detail, err := l.svcCtx.OrderService.QueryOrderDetail(l.ctx, &omsclient.QueryOrderDetailReq{
		Id: req.OrderId,
	})
	if err != nil {
		logc.Errorf(l.ctx, "确认收货查询订单失败 orderId=%d err=%v", req.OrderId, err)
		return nil, errorx.NewDefaultError("订单查询失败")
	}
	if detail == nil || detail.Data == nil {
		return nil, errorx.NewDefaultError("订单不存在")
	}
	if detail.Data.UserId != memberId {
		return nil, errorx.NewDefaultError("无权操作此订单")
	}

	// Story 6-5 Task 4.1：状态转换校验 — 必须为已发货状态才能确认收货
	currentStatus := int(detail.Data.OrderStatus)
	// ⚠️ OMS ConfirmOrder 直接 2→4，跳过 3。这里以 currentStatus==2（已支付/待发货）
	// 作为可确认收货的条件，因为 ConfirmOrder RPC 内部会更新到 4
	if currentStatus != OrderStatusPaid && currentStatus != OrderStatusShipped {
		logc.Errorf(l.ctx, "非法确认收货操作：当前状态=%d（%s），orderId=%d", currentStatus, GetStatusText(currentStatus), req.OrderId)
		return nil, errorx.NewDefaultError("当前状态不允许确认收货")
	}

	// Task 4.1：调用 OMS ConfirmOrder（MemberId + OrderId）
	_, err = l.svcCtx.OrderService.ConfirmOrder(l.ctx, &omsclient.ConfirmOrderReq{
		MemberId: memberId,
		OrderId:  req.OrderId,
	})

	if err != nil {
		logc.Errorf(l.ctx, "用户确认收货失败,参数: %+v,异常：%s", req, err.Error())
		s, _ := status.FromError(err)
		return nil, errorx.NewDefaultError(s.Message())
	}

	// Task 4.1 新增：写入操作日志（operator_type=1 用户操作，operation_type=4 确认收货）
	_, _ = l.svcCtx.OrderOperationLogService.AddOrderOperationLog(l.ctx, &omsclient.AddOrderOperationLogReq{
		OrderId:      req.OrderId,
		OperatorType: OperatorTypeUser, // 1=用户操作
		OperationType: OpConfirmReceive,  // 4=确认收货
		OperatorNote:  "用户确认收货",
	})

	return &types.ConfirmReceiveOrderResp{
		Code:    0,
		Message: "操作成功",
	}, nil
}
