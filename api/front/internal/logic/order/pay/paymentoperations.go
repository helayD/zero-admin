package pay

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/feihua/zero-admin/api/front/internal/logic/common"
	"github.com/feihua/zero-admin/api/front/internal/logic/order/order"
	"github.com/feihua/zero-admin/api/front/internal/svc"
	"github.com/feihua/zero-admin/pkg/digitalcardmint"
	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	"github.com/feihua/zero-admin/rpc/oms/omsclient"
	"github.com/smartwalle/alipay/v3"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
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
//   - 支付成功：更新 pay_status=1（OrderPaymentService）+ order_status=2（OrderService）
//   - 新增 AddOrderOperationLog 操作日志（operator_type=2 系统操作）
//   - 幂等保护：Redis key = "pay:notify:{outTradeNo}"（已实现，保留）
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
				OrderId:       paymentList.List[0].OrderId,
				OperatorType:  order.OperatorTypeSystem, // 2=系统操作
				OperationType: order.OpPaymentSuccess,   // 2=支付订单
				OperatorNote:  fmt.Sprintf("支付宝回调支付成功 outTradeNo=%s", outTradeNo),
			})
		}

		// Step 5: 发布订单支付成功消息事件（异步非阻塞，不影响支付回调主流程）
		orderId := int64(0)
		if paymentList != nil && len(paymentList.List) > 0 {
			orderId = paymentList.List[0].OrderId
		}
		if orderId > 0 {
			go func() {
				if err := l.publishPaySuccessEvent(outTradeNo, orderId); err != nil {
					logc.Errorf(l.ctx, "AliPayNotify 发布支付成功事件失败 outTradeNo=%s orderId=%d err=%v", outTradeNo, orderId, err)
				}
			}()
		}

		// Step 6: Saga 补偿链路（Epic 7 处理，本 Story 只标记，补偿在后续 Story 实施）
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
			OrderId:       paymentList.List[0].OrderId,
			OperatorType:  order.OperatorTypeSystem, // 2=系统操作
			OperationType: order.OpPaymentFailed,    // 3=支付失败（映射到业务 OpPaymentFailed）
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
	http.Error(writer, "wechat pay callback not configured", http.StatusNotImplemented)
	l.Logger.Errorf("WechatNotify 尚未接入微信支付 APIv3 回调验签与状态同步")
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

// SimulatePaySuccess 模拟支付成功（测试专用）
// Story 10.6: 测试环境未对接真实支付，通过此方法直接触发订单支付成功及后续履约分叉
func (l *PaymentOperationsUtils) SimulatePaySuccess(outTradeNo string) error {
	l.Logger.Infof("SimulatePaySuccess 开始模拟支付, outTradeNo=%s", outTradeNo)

	// Step 1: 查询支付记录获取 payment_id
	paymentList, err := l.svcCtx.OrderPaymentService.QueryOrderPaymentList(l.ctx, &omsclient.QueryOrderPaymentListReq{
		OrderNo: outTradeNo,
	})
	var paymentId int64
	if err == nil && paymentList != nil && len(paymentList.List) > 0 {
		paymentId = paymentList.List[0].Id
	}

	// Step 2: 更新 pay_status=1
	if paymentId > 0 {
		_, err = l.svcCtx.OrderPaymentService.UpdateOrderPaymentStatus(l.ctx, &omsclient.UpdateOrderPaymentStatusReq{
			Ids:       []int64{paymentId},
			PayStatus: order.PayStatusSuccess, // 1=支付成功
		})
		if err != nil {
			l.Logger.Errorf("SimulatePaySuccess 更新支付状态失败 outTradeNo=%s err=%v", outTradeNo, err)
			return fmt.Errorf("更新支付状态失败: %w", err)
		}
		l.Logger.Infof("SimulatePaySuccess 更新支付状态成功 outTradeNo=%s paymentId=%d", outTradeNo, paymentId)
	}

	// Step 3: 更新 order_status=2（已支付）
	// 先查询订单获取 ID
	var orderId int64
	if paymentList != nil && len(paymentList.List) > 0 {
		orderId = paymentList.List[0].OrderId
	}
	if orderId == 0 {
		// 通过 OrderNo 查询订单
		orderResp, queryErr := l.svcCtx.OrderService.QueryOrderList(l.ctx, &omsclient.QueryOrderListReq{
			OrderNo:  outTradeNo,
			PageNum:  1,
			PageSize: 1,
		})
		if queryErr == nil && orderResp != nil && len(orderResp.List) > 0 {
			orderId = orderResp.List[0].Id
		}
	}
	if orderId == 0 {
		return fmt.Errorf("无法获取订单ID, outTradeNo=%s", outTradeNo)
	}

	_, err = l.svcCtx.OrderService.UpdateOrder(l.ctx, &omsclient.UpdateOrderReq{
		Id:          orderId,
		OrderNo:     outTradeNo,
		OrderStatus: order.OrderStatusPaid, // 2=已支付
	})
	if err != nil {
		l.Logger.Errorf("SimulatePaySuccess 更新订单状态失败 outTradeNo=%s err=%v", outTradeNo, err)
		// 回滚支付状态
		if paymentId > 0 {
			_, _ = l.svcCtx.OrderPaymentService.UpdateOrderPaymentStatus(l.ctx, &omsclient.UpdateOrderPaymentStatusReq{
				Ids:       []int64{paymentId},
				PayStatus: order.PayStatusPending, // 0=待支付
			})
		}
		return fmt.Errorf("更新订单状态失败: %w", err)
	}
	l.Logger.Infof("SimulatePaySuccess 更新订单状态成功 outTradeNo=%s", outTradeNo)

	// Step 4: 写入操作日志
	if orderId > 0 {
		_, _ = l.svcCtx.OrderOperationLogService.AddOrderOperationLog(l.ctx, &omsclient.AddOrderOperationLogReq{
			OrderId:       orderId,
			OperatorType:  order.OperatorTypeSystem,
			OperationType: order.OpPaymentSuccess,
			OperatorNote:  fmt.Sprintf("模拟支付成功 outTradeNo=%s", outTradeNo),
		})
	}

	// Step 5: 发布支付成功 MQ 事件，触发履约分叉（Story 10.6）
	if orderId > 0 {
		eventCtx, loadErr := l.loadPaidOrderEventContext(outTradeNo, orderId)
		if loadErr != nil {
			l.Logger.Errorf("SimulatePaySuccess 获取订单上下文失败 outTradeNo=%s orderId=%d err=%v", outTradeNo, orderId, loadErr)
			return fmt.Errorf("获取订单上下文失败: %w", loadErr)
		}
		if fulfillErr := l.ensurePaidOrderPurchaseAssets(eventCtx); fulfillErr != nil {
			l.Logger.Errorf("SimulatePaySuccess 提货卡履约失败 outTradeNo=%s orderId=%d err=%v", outTradeNo, orderId, fulfillErr)
			// Story 10.6 Fix #10: 履约失败必须把订单和支付状态回滚到待支付，否则订单
			// 会停在 "已支付但卡片未建账" 的不一致状态，且 reconcile 兜底也只能补卡片，
			// 无法补订单状态。
			if paymentId > 0 {
				if _, rollbackErr := l.svcCtx.OrderPaymentService.UpdateOrderPaymentStatus(l.ctx, &omsclient.UpdateOrderPaymentStatusReq{
					Ids:       []int64{paymentId},
					PayStatus: order.PayStatusPending,
				}); rollbackErr != nil {
					l.Logger.Errorf("SimulatePaySuccess 履约失败后回滚支付状态失败 outTradeNo=%s err=%v", outTradeNo, rollbackErr)
				}
			}
			if _, rollbackErr := l.svcCtx.OrderService.UpdateOrder(l.ctx, &omsclient.UpdateOrderReq{
				Id:          orderId,
				OrderNo:     outTradeNo,
				OrderStatus: order.OrderStatusPendingPayment,
			}); rollbackErr != nil {
				l.Logger.Errorf("SimulatePaySuccess 履约失败后回滚订单状态失败 outTradeNo=%s orderId=%d err=%v", outTradeNo, orderId, rollbackErr)
			}
			return fmt.Errorf("提货卡履约失败: %w", fulfillErr)
		}
		if publishErr := l.publishPaySuccessEventWithContext(outTradeNo, eventCtx); publishErr != nil {
			l.Logger.Errorf("SimulatePaySuccess 发布支付成功事件失败 outTradeNo=%s orderId=%d err=%v", outTradeNo, orderId, publishErr)
		} else {
			l.Logger.Infof("SimulatePaySuccess 已发布支付成功事件, outTradeNo=%s orderId=%d", outTradeNo, orderId)
		}
	}

	l.Logger.Infof("SimulatePaySuccess 模拟支付完成, outTradeNo=%s", outTradeNo)
	return nil
}

// formatAmount 将元转为分（字符串），微信支付接口要求金额单位为分
func formatAmount(yuan float64) string {
	return fmt.Sprintf("%d", int(yuan*100+0.5))
}

type paidOrderEventContext struct {
	OrderID    int64  `gorm:"column:id"`
	OrderNo    string `gorm:"column:order_no"`
	MemberID   int64  `gorm:"column:member_id"`
	PlatformID int64  `gorm:"column:platform_id"`
	TenantID   int64  `gorm:"column:tenant_id"`
	MerchantID int64  `gorm:"column:merchant_id"`
}

func (c paidOrderEventContext) normalized() paidOrderEventContext {
	if c.PlatformID <= 0 {
		c.PlatformID = pkgscope.DefaultPlatformID
	}
	return c
}

func (c paidOrderEventContext) governanceScope() pkgscope.GovernanceScope {
	return pkgscope.DefaultScope(c.PlatformID, c.TenantID, c.MerchantID)
}

func (l *PaymentOperationsUtils) loadPaidOrderEventContext(outTradeNo string, orderId int64) (paidOrderEventContext, error) {
	outTradeNo = strings.TrimSpace(outTradeNo)
	if l == nil || l.svcCtx == nil {
		return paidOrderEventContext{}, errors.New("服务上下文未初始化")
	}

	if l.svcCtx.DB != nil {
		var row paidOrderEventContext
		q := l.svcCtx.DB.WithContext(l.ctx).
			Table("oms_order_main").
			Select("id, order_no, user_id AS member_id, platform_id, tenant_id, merchant_id").
			Where("is_deleted = 0")
		if orderId > 0 {
			q = q.Where("id = ?", orderId)
		} else if outTradeNo != "" {
			q = q.Where("order_no = ?", outTradeNo)
		} else {
			return paidOrderEventContext{}, errors.New("订单ID和订单号不能同时为空")
		}
		if err := q.Take(&row).Error; err == nil {
			return row.normalized(), nil
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return paidOrderEventContext{}, err
		}
	}

	fallback := l.fallbackPaidOrderEventContext(outTradeNo, orderId)
	if fallback.OrderID > 0 {
		return fallback.normalized(), nil
	}
	return paidOrderEventContext{}, fmt.Errorf("无法获取订单上下文, outTradeNo=%s orderId=%d", outTradeNo, orderId)
}

func (l *PaymentOperationsUtils) fallbackPaidOrderEventContext(outTradeNo string, orderId int64) paidOrderEventContext {
	current := common.ResolveEffectiveGovernanceScope(l.ctx)
	memberID, _ := common.GetMemberId(l.ctx)
	eventCtx := paidOrderEventContext{
		OrderID:    orderId,
		OrderNo:    outTradeNo,
		MemberID:   memberID,
		PlatformID: current.PlatformID,
		TenantID:   current.TenantID,
		MerchantID: current.MerchantID,
	}
	if orderId > 0 || strings.TrimSpace(outTradeNo) == "" || l.svcCtx == nil || l.svcCtx.OrderService == nil {
		return eventCtx.normalized()
	}

	resp, err := l.svcCtx.OrderService.QueryOrderList(l.ctx, &omsclient.QueryOrderListReq{
		OrderNo:  outTradeNo,
		PageNum:  1,
		PageSize: 1,
		Scope: &omsclient.GovernanceScope{
			ScopeType:  current.ScopeType,
			PlatformId: current.PlatformID,
			TenantId:   current.TenantID,
			MerchantId: current.MerchantID,
		},
	})
	if err == nil && resp != nil && len(resp.List) > 0 {
		eventCtx.OrderID = resp.List[0].Id
		eventCtx.OrderNo = resp.List[0].OrderNo
		if resp.List[0].UserId > 0 {
			eventCtx.MemberID = resp.List[0].UserId
		}
	}
	return eventCtx.normalized()
}

func (l *PaymentOperationsUtils) ensurePaidOrderPurchaseAssets(eventCtx paidOrderEventContext) error {
	if eventCtx.OrderID <= 0 {
		return errors.New("订单ID不能为空")
	}
	if l == nil || l.svcCtx == nil {
		return errors.New("服务上下文未初始化")
	}

	cardMintService := l.svcCtx.CardMintService
	if (cardMintService == nil || cardMintService.DB == nil) && l.svcCtx.DB != nil {
		cardMintService = digitalcardmint.NewService(l.svcCtx.DB, l.svcCtx.RabbitMQ, nil)
	}
	if cardMintService == nil || cardMintService.DB == nil {
		return errors.New("提货卡服务未初始化")
	}

	result, err := cardMintService.EnsurePaidOrderPurchaseAssets(l.ctx, digitalcardmint.EnsurePaidOrderPurchaseAssetsInput{
		OrderID:      eventCtx.OrderID,
		MemberID:     eventCtx.MemberID,
		PlatformID:   eventCtx.PlatformID,
		TenantID:     eventCtx.TenantID,
		MerchantID:   eventCtx.MerchantID,
		EventID:      fmt.Sprintf("order-paid-%d", eventCtx.OrderID),
		TraceID:      fmt.Sprintf("order-paid-%d", eventCtx.OrderID),
		OperatorType: "system",
	})
	if err != nil {
		return err
	}
	if result != nil && result.ProcessedCount > 0 {
		l.Logger.Infof("提货卡履约完成 orderId=%d count=%d", eventCtx.OrderID, result.ProcessedCount)
	}
	return nil
}

// publishPaySuccessEvent 发布订单支付成功消息事件（异步 goroutine，不阻塞支付回调主流程）
// Story 8-1 Fix #1: 修复支付成功消息缺失
func (l *PaymentOperationsUtils) publishPaySuccessEvent(outTradeNo string, orderId int64) error {
	eventCtx, err := l.loadPaidOrderEventContext(outTradeNo, orderId)
	if err != nil {
		return err
	}
	return l.publishPaySuccessEventWithContext(outTradeNo, eventCtx)
}

func (l *PaymentOperationsUtils) publishPaySuccessEventWithContext(outTradeNo string, eventCtx paidOrderEventContext) error {
	eventCtx = eventCtx.normalized()
	current := eventCtx.governanceScope()
	orderNo := firstNonEmptyString(eventCtx.OrderNo, outTradeNo)
	msgEvent := map[string]any{
		"eventId":    fmt.Sprintf("order-paid-%d", eventCtx.OrderID),
		"traceId":    fmt.Sprintf("order-paid-%d", eventCtx.OrderID),
		"scopeType":  current.ScopeType,
		"platformId": eventCtx.PlatformID,
		"tenantId":   eventCtx.TenantID,
		"merchantId": eventCtx.MerchantID,
		"actorId":    eventCtx.MemberID,
		"entityId":   eventCtx.OrderID,
		"action":     "paid",
		"version":    "v1",
		"data": map[string]any{
			"orderNo":     orderNo,
			"messageType": 2,
			"title":       "支付成功",
			"content":     fmt.Sprintf("您的订单（%s）已支付成功，感谢您的购买！", orderNo),
			"linkType":    "order",
			"linkId":      fmt.Sprintf("%d", eventCtx.OrderID),
		},
	}

	body, err := json.Marshal(msgEvent)
	if err != nil {
		return fmt.Errorf("支付成功事件序列化失败: %w", err)
	}

	if l == nil || l.svcCtx == nil || l.svcCtx.RabbitMQ == nil {
		return errors.New("RabbitMQ 未配置")
	}
	if err := l.svcCtx.RabbitMQ.SendMessage("order.event.exchange", "direct",
		"order.pay.queue", "order.paid.key", body); err != nil {
		return err
	}
	logc.Infof(l.ctx, "publishPaySuccessEvent 发送支付成功消息 outTradeNo=%s orderId=%d", orderNo, eventCtx.OrderID)
	return nil
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
