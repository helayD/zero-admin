package pay

import (
	"context"
	"fmt"
	"time"

	"github.com/feihua/zero-admin/api/front/internal/logic/common"
	"github.com/feihua/zero-admin/api/front/internal/middleware"
	"github.com/feihua/zero-admin/api/front/internal/svc"
	"github.com/feihua/zero-admin/api/front/internal/types"
	"github.com/feihua/zero-admin/rpc/oms/omsclient"
	"github.com/feihua/zero-admin/pkg/errorx"
	"github.com/zeromicro/go-zero/core/logx"
)

const (
	// PayIdempotencyPrefix Redis 幂等键前缀（支付发起）
	PayIdempotencyPrefix = "pay:idempotent:"
	// PayIdempotencyTTL 支付幂等 TTL=60s
	PayIdempotencyTTL = 60 * time.Second
)

// OrderPayLogic 预下单
type OrderPayLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewOrderPayLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OrderPayLogic {
	return &OrderPayLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// OrderPay 预下单（Task 1: 真实金额 + 状态校验 + PayType 路由 + 幂等保护）
func (l *OrderPayLogic) OrderPay(req *types.OrderPayReq) (resp *types.OrderPayResp, err error) {
	// 0. 支付方式校验（99=模拟支付，测试专用）
	if req.PayType != 1 && req.PayType != 2 && req.PayType != 99 {
		return &types.OrderPayResp{Code: 1, Message: "支付方式无效，仅支持支付宝(1)、微信(2)或模拟支付(99)"}, nil
	}

	// 1. 获取会员ID
	memberId, err := common.GetMemberId(l.ctx)
	if err != nil {
		return nil, err
	}

	// 2. 幂等保护：同一 OrderId 60s 内返回已有支付参数
	idempotencyKey := fmt.Sprintf("%s%d", PayIdempotencyPrefix, req.OrderId)
	if hit, _ := l.checkIdempotent(idempotencyKey); hit {
		return &types.OrderPayResp{Code: 0, Message: "预下单成功"}, nil
	}

	// 3. 查询 OMS 订单
	orderInfo, err := l.svcCtx.OrderService.QueryOrderDetail(l.ctx, &omsclient.QueryOrderDetailReq{
		Id:     req.OrderId,
		UserId: memberId,
	})
	if err != nil {
		return nil, errorx.NewDefaultError("OMS_ORDER_QUERY_FAILED")
	}
	if orderInfo.Data == nil {
		return &types.OrderPayResp{Code: 1, Message: "订单不存在"}, nil
	}

	// 4. 订单状态前置校验（OMS order_status: 1=待支付, 2=已支付, 5=已取消）
	omsStatus := orderInfo.Data.OrderStatus
	switch omsStatus {
	case 2: // 已支付
		return &types.OrderPayResp{Code: 1, Message: "订单已支付，请勿重复操作"}, nil
	case 3, 4: // 已发货、已完成
		return &types.OrderPayResp{Code: 1, Message: "订单已发货或已完成，无法发起支付"}, nil
	case 5: // 已取消
		return &types.OrderPayResp{Code: 1, Message: "订单已取消，无法发起支付"}, nil
	case 6, 7: // 已退款、售后中
		return &types.OrderPayResp{Code: 1, Message: "订单已退款或处于售后中，无法发起支付"}, nil
	case 1: // 待支付（正常，继续处理）
	default:
		return &types.OrderPayResp{Code: 1, Message: "订单状态异常，无法发起支付"}, nil
	}

	// 5. 提取真实支付金额（OMS 以元为单位）
	payAmountYuan := orderInfo.Data.PayAmount
	if payAmountYuan <= 0 {
		l.Logger.Errorf("订单 %d 金额异常: %f", req.OrderId, payAmountYuan)
		return &types.OrderPayResp{Code: 1, Message: "订单金额异常"}, nil
	}
	payAmountStr := fmt.Sprintf("%.2f", payAmountYuan)
	orderSubject := fmt.Sprintf("九克城订单-%s", orderInfo.Data.OrderNo)
	outTradeNo := orderInfo.Data.OrderNo

	// 6. 根据 PayType 调用对应支付渠道
	var payParam string
	operationsUtils := NewPaymentOperationsUtils(l.ctx, l.svcCtx)
	switch req.PayType {
	case 1: // 支付宝
		payParam, err = operationsUtils.TradeAppPay(outTradeNo, payAmountStr, orderSubject)
	case 2: // 微信
		payParam, err = operationsUtils.TradeAppPayWechat(outTradeNo, payAmountStr, orderSubject)
	case 99: // 模拟支付（测试专用，Story 10.6）
		if err := operationsUtils.SimulatePaySuccess(outTradeNo); err != nil {
			l.Logger.Errorf("模拟支付失败 orderId=%d err=%v", req.OrderId, err)
			return &types.OrderPayResp{Code: 1, Message: "模拟支付失败: " + err.Error()}, nil
		}
		_ = middleware.MarkCompleted(l.ctx, l.svcCtx.Redis, idempotencyKey, req.OrderId)
		return &types.OrderPayResp{Code: 0, Message: "模拟支付成功", Data: "simulate_pay_success"}, nil
	}

	if err != nil {
		l.Logger.Errorf("支付发起失败 orderId=%d payType=%d err=%v", req.OrderId, req.PayType, err)
		_ = middleware.MarkFailed(l.ctx, l.svcCtx.Redis, idempotencyKey, "OMS_ORDER_PAY_FAILED", "支付发起失败")
		return &types.OrderPayResp{Code: 1, Message: "支付发起失败，请稍后重试"}, nil
	}
	if payParam == "" {
		return &types.OrderPayResp{Code: 1, Message: "支付参数为空，请检查支付配置"}, nil
	}

	// 7. 标记幂等键为完成（TTL=24h）
	_ = middleware.MarkCompleted(l.ctx, l.svcCtx.Redis, idempotencyKey, req.OrderId)

	// 8. 返回唤起客户端的支付参数
	return &types.OrderPayResp{Code: 0, Message: "预下单成功", Data: payParam}, nil
}

// checkIdempotent 检查幂等键，返回 (是否命中, 已缓存的支付参数)
func (l *OrderPayLogic) checkIdempotent(idempotencyKey string) (bool, string) {
	result, hit, err := middleware.CheckAndSetProcessing(l.ctx, l.svcCtx.Redis, idempotencyKey)
	if err != nil {
		l.Logger.Errorf("Redis 幂等检查异常: %v", err)
		return false, ""
	}
	if !hit {
		return false, ""
	}
	switch result.State {
	case middleware.StateProcessing:
		return false, "" // 另一个请求正在处理
	case middleware.StateCompleted:
		return true, ""   // 已完成
	case middleware.StateFailed:
		return true, ""   // 失败也放行，让用户重试
	}
	return false, ""
}
