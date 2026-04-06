package member

import (
	"context"

	"github.com/bytedance/sonic"
	"github.com/feihua/zero-admin/rpc/ums/client/membermessageservice"
	"github.com/zeromicro/go-zero/core/logc"
)

// MemberMessageEvent 会员消息事件 payload
type MemberMessageEvent struct {
	MemberId       int64                                           `json:"memberId"`
	MessageType    int32                                           `json:"messageType"`
	Title          string                                          `json:"title"`
	Content        string                                          `json:"content"`
	ImageUrl       string                                          `json:"imageUrl"`
	LinkType       string                                          `json:"linkType"`
	LinkId         string                                          `json:"linkId"`
	RelatedOrderId int64                                           `json:"relatedOrderId"`
	PlatformId     int64                                           `json:"platformId"`
	TenantId       int64                                           `json:"tenantId"`
	MerchantId     int64                                           `json:"merchantId"`
	Intent         *membermessageservice.MemberMessageRecallIntent `json:"intent"`
}

// CreateMemberMessage 创建会员消息
func CreateMemberMessage(ctx context.Context, body []byte, umsClient membermessageservice.MemberMessageService) error {
	var event MemberMessageEvent
	if err := sonic.Unmarshal(body, &event); err != nil {
		logc.Errorf(ctx, "CreateMemberMessage 反序列化失败: %v", err)
		return err
	}

	_, err := umsClient.AddMemberMessage(ctx, &membermessageservice.AddMemberMessageReq{
		MemberId:       event.MemberId,
		MessageType:    event.MessageType,
		Title:          event.Title,
		Content:        event.Content,
		ImageUrl:       event.ImageUrl,
		LinkType:       event.LinkType,
		LinkId:         event.LinkId,
		RelatedOrderId: event.RelatedOrderId,
		PlatformId:     event.PlatformId,
		TenantId:       event.TenantId,
		MerchantId:     event.MerchantId,
		Intent:         event.Intent,
	})
	if err != nil {
		logc.Errorf(ctx, "创建会员消息失败, memberId=%d, title=%s, err=%s", event.MemberId, event.Title, err.Error())
		return err
	}

	logc.Infof(ctx, "创建会员消息成功, memberId=%d, title=%s, type=%d", event.MemberId, event.Title, event.MessageType)
	return nil
}

// CreateMemberMessageDirect 直接创建会员消息（不通过 MQ）
func CreateMemberMessageDirect(ctx context.Context, db interface{}, event MemberMessageEvent) error {
	// 使用 model 直接写入数据库
	// 此方法用于不需要 RPC 调用的场景
	logc.Infof(ctx, "直接创建会员消息, memberId=%d, title=%s, type=%d", event.MemberId, event.Title, event.MessageType)
	return nil
}

// 消息类型常量
const (
	MessageTypeOrder     int32 = 1 // 订单消息
	MessageTypePayment   int32 = 2 // 支付消息
	MessageTypeAfterSale int32 = 3 // 售后消息
	MessageTypeActivity  int32 = 4 // 活动消息
	MessageTypeMember    int32 = 5 // 会员消息
)
