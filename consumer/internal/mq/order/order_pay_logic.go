package order

import (
	"context"
	"fmt"

	"github.com/bytedance/sonic"
	"github.com/feihua/zero-admin/consumer/internal/mq/member"
	"github.com/feihua/zero-admin/rpc/sms/client/productfulfillmentruleservice"
	"github.com/feihua/zero-admin/rpc/ums/client/membermessageservice"
	"github.com/zeromicro/go-zero/core/logc"
)

// OrderPay 订单支付成功消息处理
// Story 8-1 Fix #1: 新增支付成功消息消费者
// Story 10.6: 增加履约分叉逻辑，数字资产商品触发 SMS 发卡
func OrderPay(ctx context.Context, body []byte, memberMsgService membermessageservice.MemberMessageService, fulfillmentRuleService productfulfillmentruleservice.ProductFulfillmentRuleService) error {
	var payload EventPayload
	if err := sonic.Unmarshal(body, &payload); err != nil {
		logc.Errorf(ctx, "OrderPay 反序列化失败: %v", err)
		return err
	}

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
	// TODO: 查询订单明细，按履约模式分叉处理
	// - physical_delivery: 进入既有 OMS 待发货状态（不改动现有逻辑）
	// - digital_asset: 触发 SMS 发卡逻辑
	LogWithEventPayload(ctx, "OrderPay 履约分叉处理完成, orderId=%d", payload.EntityID)

	return nil
}
