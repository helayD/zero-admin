package order

import (
	"context"
	"fmt"

	"github.com/bytedance/sonic"
	"github.com/feihua/zero-admin/consumer/internal/mq/member"
	"github.com/feihua/zero-admin/pkg/digitalcardmint"
	"github.com/feihua/zero-admin/rpc/ums/client/membermessageservice"
	"github.com/zeromicro/go-zero/core/logc"
	"gorm.io/gorm"
)

// OrderPay 订单支付成功消息处理
// Story 8-1 Fix #1: 新增支付成功消息消费者
// Story 10.6: 增加履约分叉逻辑，提货卡商品触发 SMS 发卡
func OrderPay(ctx context.Context, body []byte, memberMsgService membermessageservice.MemberMessageService, cardMintService *digitalcardmint.Service, db *gorm.DB) error {
	var payload EventPayload
	if err := sonic.Unmarshal(body, &payload); err != nil {
		logc.Errorf(ctx, "OrderPay 反序列化失败: %v", err)
		return err
	}
	payload.NormalizeLegacyPaymentPayload()

	ctx = payload.ToContext(ctx)
	LogWithEventPayload(ctx, "OrderPay 收到订单支付成功事件, entityId=%d, action=%s", payload.EntityID, payload.Action)

	orderNo := ""
	if v, ok := payload.Data["orderNo"]; ok {
		orderNo, _ = v.(string)
	}

	// 1. 发送支付成功消息通知
	msgEvent := member.MemberMessageEvent{
		MemberId:    payload.ActorID,
		MessageType: member.MessageTypePayment,
		Title:       "支付成功",
		Content:     fmt.Sprintf("您的订单（%s）已支付成功，感谢您的购买！", orderNo),
		LinkType:    "order",
		LinkId:      payload.EntityIDToString(),
		PlatformId:  payload.PlatformID,
		TenantId:    payload.TenantID,
		MerchantId:  payload.MerchantID,
	}

	body2, err := sonic.Marshal(msgEvent)
	if err != nil {
		LogWithEventPayload(ctx, "OrderPay 序列化消息事件失败: %v", err)
		return nil
	}

	if err := member.CreateMemberMessage(ctx, body2, memberMsgService); err != nil {
		LogWithEventPayload(ctx, "OrderPay 发送支付成功消息失败: %v", err)
		// 不返回错误，继续处理履约分叉
	}

	// 2. 履约分叉逻辑（Story 10.6）
	if err := processPaidOrderDigitalAssets(ctx, &payload, cardMintService, db); err != nil {
		LogWithEventPayload(ctx, "OrderPay 提货卡履约分叉失败, orderId=%d, err=%v", payload.EntityID, err)
		return err
	}
	LogWithEventPayload(ctx, "OrderPay 履约分叉处理完成, orderId=%d", payload.EntityID)

	return nil
}

func processPaidOrderDigitalAssets(ctx context.Context, payload *EventPayload, cardMintService *digitalcardmint.Service, db *gorm.DB) error {
	if payload == nil || payload.EntityID <= 0 {
		return nil
	}
	if cardMintService == nil && db != nil {
		cardMintService = digitalcardmint.NewService(db, nil, nil)
	}
	if cardMintService == nil {
		return nil
	}

	result, err := cardMintService.EnsurePaidOrderPurchaseAssets(ctx, digitalcardmint.EnsurePaidOrderPurchaseAssetsInput{
		OrderID:      payload.EntityID,
		MemberID:     payload.ActorID,
		PlatformID:   payload.PlatformID,
		TenantID:     payload.TenantID,
		MerchantID:   payload.MerchantID,
		EventID:      payload.EventID,
		TraceID:      payload.TraceID,
		OperatorType: "system",
	})
	if err != nil {
		return err
	}
	if result != nil && result.ProcessedCount > 0 {
		logc.Infof(ctx, "OrderPay 已创建或确认提货卡资产, orderId=%d, count=%d, traceId=%s", payload.EntityID, result.ProcessedCount, payload.TraceID)
	}
	return nil
}
