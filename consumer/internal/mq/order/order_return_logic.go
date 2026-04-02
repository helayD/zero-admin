package order

import (
	"context"
	"fmt"

	"github.com/bytedance/sonic"
	"github.com/feihua/zero-admin/consumer/internal/mq/member"
	"github.com/feihua/zero-admin/rpc/ums/client/membermessageservice"
	"github.com/zeromicro/go-zero/core/logc"
)

// OrderReturn 售后申请事件处理
func OrderReturn(ctx context.Context, body []byte, memberMsgService membermessageservice.MemberMessageService) error {
	var payload EventPayload
	if err := sonic.Unmarshal(body, &payload); err != nil {
		logc.Errorf(ctx, "OrderReturn 反序列化失败: %v", err)
		return err
	}

	ctx = payload.ToContext(ctx)
	LogWithEventPayload(ctx, "OrderReturn 收到售后申请事件, entityId=%d, action=%s", payload.EntityID, payload.Action)

	orderNo := ""
	if v, ok := payload.Data["orderNo"]; ok {
		orderNo, _ = v.(string)
	}

	msgEvent := member.MemberMessageEvent{
		MemberId:    payload.ActorID,
		MessageType: member.MessageTypeAfterSale,
		Title:       "售后申请已提交",
		Content:     fmt.Sprintf("您的售后申请（订单号: %s）已提交，请等待处理", orderNo),
		LinkType:    "order",
		LinkId:      fmt.Sprintf("%d", payload.EntityID),
		PlatformId:  payload.PlatformID,
		TenantId:    payload.TenantID,
		MerchantId:  payload.MerchantID,
	}

	body2, err := sonic.Marshal(msgEvent)
	if err != nil {
		LogWithEventPayload(ctx, "OrderReturn 序列化消息事件失败: %v", err)
		return nil
	}

	return member.CreateMemberMessage(ctx, body2, memberMsgService)
}
