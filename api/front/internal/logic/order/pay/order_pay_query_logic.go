package pay

import (
	"context"
	"time"

	"github.com/feihua/zero-admin/api/front/internal/logic/common"
	orderlogic "github.com/feihua/zero-admin/api/front/internal/logic/order/order"
	"github.com/feihua/zero-admin/api/front/internal/svc"
	"github.com/feihua/zero-admin/api/front/internal/types"
	"github.com/feihua/zero-admin/rpc/oms/omsclient"
	"github.com/zeromicro/go-zero/core/logx"
)

const (
	// DefaultPayTimeoutMinutes 默认支付超时时间（分钟）
	DefaultPayTimeoutMinutes = 30
)

// OrderPayQueryLogic 支付状态查询
type OrderPayQueryLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewOrderPayQueryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OrderPayQueryLogic {
	return &OrderPayQueryLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// OrderPayQuery 支付状态查询（Task 2: OMS状态 + 第三方状态融合）
func (l *OrderPayQueryLogic) OrderPayQuery(req *types.OrderPayQueryReq) (resp *types.OrderPayQueryResp, err error) {
	// 1. 获取会员ID
	memberId, err := common.GetMemberId(l.ctx)
	if err != nil {
		return nil, err
	}

	// 2. 查询 OMS 订单
	orderInfo, err := l.svcCtx.OrderService.QueryOrderDetail(l.ctx, &omsclient.QueryOrderDetailReq{
		Id:     req.OrderId,
		UserId: memberId,
	})
	if err != nil || orderInfo.Data == nil {
		return payQueryResp(1, "查询订单失败", "", 0, 0, 0, nil), nil
	}

	order := orderInfo.Data

	// 3. OMS 状态 → 前端映射（OMS: 1=待支付,2=已支付,5=已取消）
	// 前端: orderStatus 0=待支付 1=已支付 2=已取消
	var frontendOrderStatus int64
	var frontendPayStatus int64
	var message string

	switch order.OrderStatus {
	case 2: // 已支付
		frontendOrderStatus = 1
		frontendPayStatus = 1
		message = "订单已支付"
	case 5: // 已取消
		frontendOrderStatus = 2
		frontendPayStatus = 0
		message = "订单已取消"
	default:
		// 1=待支付 或其他状态，走第三方查询
		frontendOrderStatus = 0
		frontendPayStatus = 0
		message = "待支付"

		// 4. OMS 仍待支付，查询第三方支付状态
		operationsUtils := NewPaymentOperationsUtils(l.ctx, l.svcCtx)
		_, tradeStatus, thirdErr := operationsUtils.TradeQuery(order.OrderNo)

		if thirdErr == nil && tradeStatus == 1 {
			// 第三方已支付，触发 Saga 补偿：更新 OMS 状态
			operationsUtils.UpdatePaidStatus(order.OrderNo)
			frontendOrderStatus = 1
			frontendPayStatus = 1
			message = "支付成功"
		}
	}

	// 5. 计算剩余支付时间
	expireTime := calculateExpireTime(order.CreateTime)

	snapshot, snapshotErr := orderlogic.NewOrderStatusService(l.ctx, l.svcCtx).GetOrderStatusSnapshot(req.OrderId, memberId)
	if snapshotErr != nil {
		return payQueryResp(0, message, "", frontendOrderStatus, frontendPayStatus, expireTime, nil), nil
	}

	return payQueryResp(0, message, "", frontendOrderStatus, frontendPayStatus, expireTime, snapshot), nil
}

// calculateExpireTime 计算剩余支付秒数
func calculateExpireTime(createTime string) int64 {
	if createTime == "" {
		return int64(DefaultPayTimeoutMinutes * 60)
	}
	parsedTime, err := time.ParseInLocation("2006-01-02 15:04:05", createTime, time.Local)
	if err != nil {
		return int64(DefaultPayTimeoutMinutes * 60)
	}
	expireAt := parsedTime.Add(time.Duration(DefaultPayTimeoutMinutes) * time.Minute)
	remaining := time.Until(expireAt).Seconds()
	if remaining < 0 {
		return 0
	}
	return int64(remaining)
}

func payQueryResp(code int64, message, data string, orderStatus, payStatus, expireTime int64, snapshot *orderlogic.OrderSnapshot) *types.OrderPayQueryResp {
	resp := &types.OrderPayQueryResp{
		Code:        code,
		Message:     message,
		Data:        data,
		OrderStatus: orderStatus,
		PayStatus:   payStatus,
		ExpireTime:  expireTime,
	}
	if snapshot != nil {
		resp.ConsistencyStage = snapshot.ConsistencyStage
		resp.ConsistencyStageText = snapshot.ConsistencyStageText
		resp.ConsistencyResult = snapshot.ConsistencyResult
		resp.ConsistencyMessage = snapshot.ConsistencyMessage
		resp.LastConsistencyAt = snapshot.LastConsistencyAt
		resp.PendingActions = snapshot.PendingActions
		resp.PendingActionsText = snapshot.PendingActionsText
		resp.AftersaleStatus = snapshot.AftersaleStatus
		resp.AftersaleStatusText = snapshot.AftersaleStatusText
		resp.ReturnId = snapshot.ReturnId
		resp.ReturnNo = snapshot.ReturnNo
	}
	return resp
}
