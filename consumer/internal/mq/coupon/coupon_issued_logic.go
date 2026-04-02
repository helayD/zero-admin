package coupon

import (
	"context"

	"github.com/bytedance/sonic"
	"github.com/feihua/zero-admin/consumer/internal/mq/member"
	"github.com/feihua/zero-admin/rpc/ums/client/membermessageservice"
	"github.com/zeromicro/go-zero/core/logc"
)

func CouponIssued(ctx context.Context, body []byte, memberMsgService membermessageservice.MemberMessageService) error {
	var payload map[string]interface{}
	if err := sonic.Unmarshal(body, &payload); err != nil {
		logc.Errorf(ctx, "CouponIssued 反序列化失败: %v", err)
		return err
	}

	memberId, _ := payload["memberId"].(float64)
	couponName, _ := payload["couponName"].(string)

	logc.Infof(ctx, "CouponIssued 收到优惠券发放事件, memberId=%d, couponName=%s", int64(memberId), couponName)

	msgEvent := member.MemberMessageEvent{
		MemberId:    int64(memberId),
		MessageType: member.MessageTypeMember,
		Title:       "优惠券到账",
		Content:     "您有一张新优惠券：「" + couponName + "」，快去使用吧！",
		LinkType:    "coupon",
		LinkId:     "",
	}

	body2, err := sonic.Marshal(msgEvent)
	if err != nil {
		logc.Errorf(ctx, "CouponIssued 序列化消息事件失败: %v", err)
		return nil
	}

	return member.CreateMemberMessage(ctx, body2, memberMsgService)
}
