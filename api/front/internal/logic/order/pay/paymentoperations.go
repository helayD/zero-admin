package pay

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/feihua/zero-admin/api/front/internal/logic/order/order"
	"github.com/feihua/zero-admin/api/front/internal/svc"
	"github.com/feihua/zero-admin/rpc/oms/omsclient"
	"github.com/smartwalle/alipay/v3"
	"github.com/zeromicro/go-zero/core/logx"
)

// PaymentOperationsUtils 支付相关工具
// 重构说明（Story 6-5）：
//  - AliPayNotify 支付成功：使用 OrderPaymentService.UpdateOrderPaymentStatus 更新 pay_status
//  - 新增 AddOrderOperationLog 操作日志记录（operator_type=2 系统操作）
//  - 幂等保护：Redis key = "pay:notify:{outTradeNo}"，TTL=24h（已实现，保留）
/*
Author: LiuFeiHua
Date: 2023/12/15 10:05
Story: 5.5 Tasks 3, 4: 微信支付支持 + 回调幂等 + WechatNotify
*/
type PaymentOperationsUtils struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPaymentOperationsUtils(ctx context.Context, svcCtx *svc.ServiceContext) *PaymentOperationsUtils {
	return &PaymentOperationsUtils{Logger: logx.WithContext(ctx), ctx: ctx, svcCtx: svcCtx}
}

// ==================== 支付宝支付（已有）====================

// TradeAppPay 支付宝 APP 支付
func (l *PaymentOperationsUtils) TradeAppPay(outTradeNo, totalAmount, subject string) (string, error) {
	var p = alipay.TradeAppPay{}
	p.NotifyURL = l.svcCtx.Config.Alipay.NotifyURL
	p.Subject = subject
	p.OutTradeNo = outTradeNo
	p.TotalAmount = totalAmount
	return l.svcCtx.AlipayClient.TradeAppPay(p)
}

// TradeQuery 查询支付宝订单
func (l *PaymentOperationsUtils) TradeQuery(outTradeNo string) (string, int64, error) {
	var p = alipay.TradeQuery{}
	p.OutTradeNo = outTradeNo
	tradeQuery, err := l.svcCtx.AlipayClient.TradeQuery(l.ctx, p)
	if err != nil {
		return outTradeNo, 0, err
	}
	if tradeQuery.IsFailure() {
		return outTradeNo, 0, errors.New("alipay trade query failed")
	}
	if tradeQuery.TradeStatus == alipay.TradeStatusSuccess {
		return outTradeNo, 1, nil
	}
	return outTradeNo, 0, nil
}

// ==================== Task 3: 微信支付 ====================

// TradeAppPayWechat 微信 APP 支付（Story 5.5 新增）
// NOTE: 微信 APP 支付需商户在微信支付平台配置 APPID 和商户号，并使用微信支付 SDK
// 此处为 stub 实现，实际需接入 wechatpay-go SDK 并配置 WechatConfig
func (l *PaymentOperationsUtils) TradeAppPayWechat(outTradeNo, totalAmount, subject string) (string, error) {
	// TODO(5.5): 接入 wechatpay-go SDK
	// 参考: https://github.com/wechatpay-apiv3/wechatpay-go
	// 实现步骤:
	// 1. 初始化微信支付客户端: wechatpay.New(...)
	// 2. 构造统信请求: clientv3.H5Client.NewClient()
	// 3. 调用 JSAPI/APP 支付接口
	// 4. 返回 mweb_url（H5支付）或 appid/prepayid（APP支付）
	l.Logger.Errorf("微信 APP 支付暂未配置，请完成微信支付商户配置后接入 wechatpay-go SDK")
	return "", errors.New("wechat pay not configured: wechatpay-go SDK integration pending")
}

// TradeQueryWechat 查询微信支付订单（Story 5.5 新增）
func (l *PaymentOperationsUtils) TradeQueryWechat(outTradeNo string) (string, int64, error) {
	// TODO(5.5): 接入 wechatpay-go SDK
	// 参考: https://github.com/wechatpay-apiv3/wechatpay-go
	// 实现步骤:
	// 1. 调用微信支付订单查询接口 GET /v3/pay/transactions/out-trade-no/{out_trade_no}
	// 2. 解析返回的 trade_state: SUCCESS=已支付, NOTPAY=未支付, CLOSED=已关闭
	l.Logger.Errorf("微信订单查询暂未配置")
	return outTradeNo, 0, errors.New("wechat trade query not configured")
}

// ==================== Task 4: 回调处理 ====================

// AliPayNotify 支付宝回调通知（Story 5.5: 幂等处理 + Saga补偿）
// 重构说明（Story 6-5 Task 5）：
//  - 支付成功：更新 pay_status=1（OrderPaymentService）+ order_status=2（OrderService）
//  - 新增 AddOrderOperationLog 操作日志（operator_type=2 系统操作）
//  - 幂等保护：Redis key = "pay:notify:{outTradeNo}"（已实现，保留）
func (l *PaymentOperationsUtils) AliPayNotify(writer http.ResponseWriter, request *http.Request) {
	if err := request.ParseForm(); err != nil {
		_, _ = writer.Write([]byte("error"))
		return
	}

	notification, err := l.svcCtx.AlipayClient.DecodeNotification(request.Form)
	if err != nil {
		_, _ = writer.Write([]byte("error"))
		return
	}

	l.WithContext(l.ctx).Infof("支付宝支付回调参数：%+v", notification)
	outTradeNo := notification.OutTradeNo
	tradeStatus := notification.TradeStatus

	if alipay.TradeStatusSuccess == tradeStatus {
		// Task 4.3 幂等处理（已实现，保留）
		if l.isPayStatusUpdated(outTradeNo) {
			l.Logger.Infof("AliPayNotify 幂等命中 outTradeNo=%s，已更新过，跳过", outTradeNo)
			_, _ = writer.Write([]byte("success"))
			return
		}

		// Story 6-5 Task 5.2: 支付成功时更新 pay_status=1 + order_status=2

		// Step 1: 查询支付记录获取 payment_id（UpdateOrderPaymentStatusReq.ids 是 payment_record.id）
		paymentList, err := l.svcCtx.OrderPaymentService.QueryOrderPaymentList(l.ctx, &omsclient.QueryOrderPaymentListReq{
			OrderNo: outTradeNo,
		})
		var paymentId int64
		if err == nil && paymentList != nil && len(paymentList.List) > 0 {
			paymentId = paymentList.List[0].Id
		}

		// Step 2: 更新 pay_status=1（OrderPaymentService.UpdateOrderPaymentStatus）
		// ⚠️ UpdateOrderPaymentStatusReq.ids = payment_record.id，不是 order_id
		if paymentId > 0 {
			_, err = l.svcCtx.OrderPaymentService.UpdateOrderPaymentStatus(l.ctx, &omsclient.UpdateOrderPaymentStatusReq{
				Ids:       []int64{paymentId},
				PayStatus: order.PayStatusSuccess, // OMS 1=支付成功
			})
			if err != nil {
				l.Logger.Errorf("AliPayNotify 更新支付状态失败 outTradeNo=%s err=%v", outTradeNo, err)
			} else {
				l.Logger.Infof("AliPayNotify 更新支付状态成功 outTradeNo=%s paymentId=%d", outTradeNo, paymentId)
			}
		}

		// Step 3: 更新 order_status=2（OrderService.UpdateOrder）
		_, err = l.svcCtx.OrderService.UpdateOrder(l.ctx, &omsclient.UpdateOrderReq{
			OrderNo:     outTradeNo,
			OrderStatus: order.OrderStatusPaid, // OMS 2=已支付
		})
		if err != nil {
			l.Logger.Errorf("AliPayNotify 更新订单状态失败 outTradeNo=%s err=%v", outTradeNo, err)
		} else {
			l.Logger.Infof("AliPayNotify 更新订单状态成功 outTradeNo=%s", outTradeNo)
		}

		// Step 4: 写入操作日志（operator_type=2 系统操作，operation_type=2 支付订单）
		if paymentId > 0 {
			_, _ = l.svcCtx.OrderOperationLogService.AddOrderOperationLog(l.ctx, &omsclient.AddOrderOperationLogReq{
				OrderId:      paymentList.List[0].OrderId,
				OperatorType: order.OperatorTypeSystem, // 2=系统操作
				OperationType: order.OpPaymentSuccess,   // 2=支付订单
				OperatorNote:  fmt.Sprintf("支付宝回调支付成功 outTradeNo=%s", outTradeNo),
			})
		}

		// Step 5: Saga 补偿链路（Epic 7 处理，本 Story 只标记，补偿在后续 Story 实施）
		// 库存扣减（Story 7-3a）：下单时已锁定，取消时释放
		// 优惠券核销（Story 7-3a）：CancelOrderResp.CouponIds 已归还
		// 积分扣减（Story 7-3a）：支付成功时 saga 确认

		l.svcCtx.AlipayClient.ACKNotification(writer)
		return
	}

	// 支付失败：更新 pay_status=2，不改变 order_status
	// Story 6-5 Task 5.3
	l.Logger.Infof("AliPayNotify 支付失败 outTradeNo=%s tradeStatus=%s", outTradeNo, tradeStatus)

	paymentList, err := l.svcCtx.OrderPaymentService.QueryOrderPaymentList(l.ctx, &omsclient.QueryOrderPaymentListReq{
		OrderNo: outTradeNo,
	})
	var paymentId int64
	if err == nil && paymentList != nil && len(paymentList.List) > 0 {
		paymentId = paymentList.List[0].Id
	}

	if paymentId > 0 {
		_, _ = l.svcCtx.OrderPaymentService.UpdateOrderPaymentStatus(l.ctx, &omsclient.UpdateOrderPaymentStatusReq{
			Ids:       []int64{paymentId},
			PayStatus: order.PayStatusFailed, // OMS 2=支付失败
		})

		// 写入操作日志
		_, _ = l.svcCtx.OrderOperationLogService.AddOrderOperationLog(l.ctx, &omsclient.AddOrderOperationLogReq{
			OrderId:      paymentList.List[0].OrderId,
			OperatorType: order.OperatorTypeSystem, // 2=系统操作
			OperationType: order.OpPaymentFailed,   // 3=支付失败（映射到业务 OpPaymentFailed）
			OperatorNote:  fmt.Sprintf("支付宝回调支付失败 outTradeNo=%s", outTradeNo),
		})
	}

	_, _ = writer.Write([]byte("success"))
}

// isPayStatusUpdated 幂等检查：查询订单是否已更新为已支付（Story 5.5 Task 4.3）
// NOTE: QueryOrderDetailReq 仅支持 Id 字段，不支持 OrderNo 查询。
// 为避免不必要的 OMS 调用，这里通过 Redis key 记录已处理的 outTradeNo。
// key = "pay:notify:{outTradeNo}"，收到回调时 SETNX，TTL=24h。
func (l *PaymentOperationsUtils) isPayStatusUpdated(outTradeNo string) bool {
	key := fmt.Sprintf("pay:notify:%s", outTradeNo)
	// 使用 SetnxEx 原子设置：若 key 不存在则设置成功（false=未处理）；若已存在则返回true（已处理）
	set, err := l.svcCtx.Redis.SetnxExCtx(l.ctx, key, "1", 86400)
	if err != nil {
		l.Logger.Errorf("Redis SetnxEx 幂等检查异常 outTradeNo=%s err=%v", outTradeNo, err)
		return false // 异常时放行，避免阻塞回调
	}
	// set=true = 设置成功（之前未处理）；set=false = key 已存在（已处理）
	return !set
}

// WechatNotify 微信支付回调通知（Story 5.5 Task 4.2 新增）
// 微信支付回调通常通过微信支付后台配置的 NotifyURL 推送
// NOTE: 需在微信支付商户平台配置回调 URL，接入微信支付 APIv3
func (l *PaymentOperationsUtils) WechatNotify(writer http.ResponseWriter, request *http.Request) {
	// TODO(5.5): 接入微信支付 APIv3 回调验签和处理
	// 参考: https://github.com/wechatpay-apiv3/wechatpay-go
	// 实现步骤:
	// 1. 从请求 Header 获取 Wechatpay-Signature, Wechatpay-Nonce, Wechatpay-Timestamp
	// 2. 构造签名串: timestamp + nonce + request_body
	// 3. 使用平台证书验签
	// 4. 解析 JSON 请求体获取 transaction_id, out_trade_no, trade_state
	// 5. trade_state == "SUCCESS" → 更新 OMS 状态（复用 AliPayNotify 逻辑）
	// 6. 返回 HTTP 200

	l.Logger.Infof("WechatNotify 收到回调")
	_, _ = writer.Write([]byte("success"))
}

// UpdatePaidStatus 支付成功后的 Saga 补偿（OMS 状态同步）
// ⚠️ 已废弃（Story 6-5 Task 5 重构）
// AliPayNotify 已内联处理 pay_status=1 + order_status=2 + 操作日志
// 此方法保留仅用于兼容外部调用，实际逻辑已迁移到 AliPayNotify
func (l *PaymentOperationsUtils) UpdatePaidStatus(outTradeNo string) {
	// Note: 当前 UpdateOrderReq 不含 pay_status
	// Saga 补偿分两步：
	// 1. 更新 orderStatus: 已有 UpdateOrder
	// 2. 更新 payStatus:  需扩展 OMS proto 或使用 UpdateOrderPaymentStatus RPC
	// 此处先用 orderStatus=2 表示已支付，payStatus 的更新见 Saga 说明
	_, err := l.svcCtx.OrderService.UpdateOrder(l.ctx, &omsclient.UpdateOrderReq{
		OrderNo:     outTradeNo,
		OrderStatus: order.OrderStatusPaid, // OMS 2=已支付
	})
	if err != nil {
		l.Logger.Errorf("UpdatePaidStatus 更新订单状态失败 outTradeNo=%s err=%v", outTradeNo, err)
	} else {
		l.Logger.Infof("UpdatePaidStatus 更新订单状态成功 outTradeNo=%s", outTradeNo)
	}
}

// formatAmount 将元转为分（字符串），微信支付接口要求金额单位为分
func formatAmount(yuan float64) string {
	return fmt.Sprintf("%d", int(yuan*100+0.5))
}
