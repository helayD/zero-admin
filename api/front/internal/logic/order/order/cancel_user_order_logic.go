package order

import (
	"context"

	"github.com/feihua/zero-admin/api/front/internal/logic/common"
	"github.com/feihua/zero-admin/pkg/errorx"
	"github.com/feihua/zero-admin/rpc/oms/omsclient"
	"github.com/feihua/zero-admin/rpc/pms/pmsclient"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"
	"github.com/feihua/zero-admin/rpc/ums/umsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"google.golang.org/grpc/status"

	"github.com/feihua/zero-admin/api/front/internal/svc"
	"github.com/feihua/zero-admin/api/front/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

// CancelUserOrderLogic 取消订单
// 重构说明（Story 6-5）：
//  - 使用 OrderStatusService 进行状态转换校验（必须为 OrderStatusPendingPayment=1）
//  - Saga 补偿链路（库存/优惠券/积分）完整保留，不通过状态机处理
//  - 新增 AddOrderOperationLog 操作日志记录
/*
Author: LiuFeiHua
Date: 2025/6/20 9:41
*/
type CancelUserOrderLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCancelUserOrderLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CancelUserOrderLogic {
	return &CancelUserOrderLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// CancelUserOrder 取消订单
func (l *CancelUserOrderLogic) CancelUserOrder(req *types.CancelUserOrderReq) (resp1 *types.CancelUserOrderResp, err error) {
	memberId, err := common.GetMemberId(l.ctx)
	if err != nil {
		return nil, err
	}

	// Story 6-5 Task 3.1：状态机校验 — 必须为待支付状态才能取消
	orderSvc := NewOrderStatusService(l.ctx, l.svcCtx)
	currentStatus, err := orderSvc.getCurrentOrderStatus(req.OrderId)
	if err != nil {
		logc.Errorf(l.ctx, "取消订单查询状态失败 orderId=%d err=%v", req.OrderId, err)
		return nil, errorx.NewDefaultError("订单查询失败")
	}
	if !IsValidTransition(currentStatus, OrderStatusCancelled) {
		logc.Errorf(l.ctx, "非法取消操作：当前状态=%d（%s），orderId=%d", currentStatus, GetStatusText(currentStatus), req.OrderId)
		return nil, errorx.NewDefaultError("当前状态不允许取消订单")
	}

	// 归属校验（双重保护）
	detail, err := l.svcCtx.OrderService.QueryOrderDetail(l.ctx, &omsclient.QueryOrderDetailReq{
		Id: req.OrderId,
	})
	if err == nil && detail != nil && detail.Data != nil && detail.Data.UserId != memberId {
		return nil, errorx.NewDefaultError("无权操作此订单")
	}

	// Task 3.1：Saga 补偿链路（第 47 行原有逻辑完整保留）
	// 编排说明：OMS CancelOrder 内部处理订单状态变更，返回锁定库存、已用优惠券、已扣积分
	// 后续 4 步按序执行，当前为串行编排，Story 7-3a 评估引入消息队列实现最终一致性
	resp, err := l.svcCtx.OrderService.CancelOrder(l.ctx, &omsclient.CancelOrderReq{
		MemberId: memberId,
		OrderId:  req.OrderId,
	})
	if err != nil {
		logc.Errorf(l.ctx, "OMS CancelOrder 失败 orderId=%d err=%v", req.OrderId, err)
		s, _ := status.FromError(err)
		return nil, errorx.NewDefaultError(s.Message())
	}

	couponIds := resp.CouponIds
	integration := resp.Integration
	stockLockData := resp.Data

	// Task 2.4：Saga 补偿链路 Step 3 — 释放库存
	var data []*pmsclient.UpdateSkuStockData
	for _, item := range stockLockData {
		data = append(data, &pmsclient.UpdateSkuStockData{
			Id:              item.ProductSkuId,
			ProductQuantity: item.ProductQuantity,
		})
	}
	_, err = l.svcCtx.ProductSkuService.ReleaseSkuStockLock(l.ctx, &pmsclient.UpdateSkuStockReq{
		Data: data,
	})
	if err != nil {
		return nil, err
	}

	// Task 2.4：Saga 补偿链路 Step 4 — 归还优惠券
	if len(couponIds) > 0 {
		_, err = l.svcCtx.CouponRecordService.UpdateCouponRecord(l.ctx, &smsclient.UpdateCouponRecordReq{
			MemberId:  memberId,
			Status:    0,
			CouponIds: couponIds,
		})
		if err != nil {
			return nil, err
		}
	}

	// Task 2.4：Saga 补偿链路 Step 5 — 返还积分
	member, _ := l.svcCtx.MemberService.QueryMemberInfoDetail(l.ctx, &umsclient.QueryMemberInfoDetailReq{MemberId: memberId})
	i := member.Points + integration
	_, err = l.svcCtx.MemberService.UpdateMemberPoints(l.ctx, &umsclient.UpdateMemberPointsReq{MemberId: memberId, Points: i})
	if err != nil {
		return nil, err
	}

	// Task 3.1 新增：写入操作日志（operator_type=1 用户操作，operation_type=5 取消订单）
	_, _ = l.svcCtx.OrderOperationLogService.AddOrderOperationLog(l.ctx, &omsclient.AddOrderOperationLogReq{
		OrderId:      req.OrderId,
		OperatorType: OperatorTypeUser, // 1=用户操作
		OperationType: OpCancel,         // 6 取消订单（复用业务 Op 常量）
		OperatorNote:  "用户取消订单",
	})

	return &types.CancelUserOrderResp{
		Code:    0,
		Message: "取消订单成功",
	}, nil
}
