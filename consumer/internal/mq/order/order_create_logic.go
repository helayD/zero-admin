package order

import (
	"context"
	"fmt"

	"github.com/bytedance/sonic"
	"github.com/feihua/zero-admin/consumer/internal/mq/member"
	"github.com/feihua/zero-admin/rpc/ums/client/membermessageservice"
	"github.com/zeromicro/go-zero/core/logc"
)

func OrderCreate(ctx context.Context, body []byte, memberMsgService membermessageservice.MemberMessageService) error {
	var payload EventPayload
	if err := sonic.Unmarshal(body, &payload); err != nil {
		logc.Errorf(ctx, "OrderCreate 反序列化失败: %v", err)
		return err
	}

	ctx = payload.ToContext(ctx)
	LogWithEventPayload(ctx, "OrderCreate 收到订单创建事件, entityId=%d, action=%s", payload.EntityID, payload.Action)

	orderNo := ""
	if v, ok := payload.Data["orderNo"]; ok {
		orderNo, _ = v.(string)
	}

	msgEvent := member.MemberMessageEvent{
		MemberId:    payload.ActorID,
		MessageType: member.MessageTypeOrder,
		Title:       "订单已提交",
		Content:     fmt.Sprintf("您的订单（%s）已成功提交，请尽快完成支付", orderNo),
		LinkType:    "order",
		LinkId:      payload.EntityIDToString(),
		PlatformId: payload.PlatformID,
		TenantId:   payload.TenantID,
		MerchantId: payload.MerchantID,
	}

	body2, err := sonic.Marshal(msgEvent)
	if err != nil {
		LogWithEventPayload(ctx, "OrderCreate 序列化消息事件失败: %v", err)
		return nil
	}

	return member.CreateMemberMessage(ctx, body2, memberMsgService)
}
