package order

import (
	"context"
	"fmt"

	"github.com/bytedance/sonic"
	"github.com/feihua/zero-admin/consumer/internal/mq/member"
	"github.com/feihua/zero-admin/rpc/ums/client/membermessageservice"
	"github.com/zeromicro/go-zero/core/logc"
)

func OrderConfirm(ctx context.Context, body []byte, memberMsgService membermessageservice.MemberMessageService) error {
	var payload EventPayload
	if err := sonic.Unmarshal(body, &payload); err != nil {
		logc.Errorf(ctx, "OrderConfirm 反序列化失败: %v", err)
		return err
	}

	ctx = payload.ToContext(ctx)
	LogWithEventPayload(ctx, "OrderConfirm 收到订单确认收货事件, entityId=%d, action=%s", payload.EntityID, payload.Action)

	orderNo := ""
	if v, ok := payload.Data["orderNo"]; ok {
		orderNo, _ = v.(string)
	}

	msgEvent := member.MemberMessageEvent{
		MemberId:    payload.ActorID,
		MessageType: member.MessageTypeOrder,
		Title:       "确认收货成功",
		Content:     fmt.Sprintf("您的订单（%s）已确认收货，感谢您的购买！", orderNo),
		LinkType:    "order",
		LinkId:      payload.EntityIDToString(),
		PlatformId: payload.PlatformID,
		TenantId:   payload.TenantID,
		MerchantId: payload.MerchantID,
	}

	body2, err := sonic.Marshal(msgEvent)
	if err != nil {
		LogWithEventPayload(ctx, "OrderConfirm 序列化消息事件失败: %v", err)
		return nil
	}

	return member.CreateMemberMessage(ctx, body2, memberMsgService)
}
