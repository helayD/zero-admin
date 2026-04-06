package membermessageservicelogic

import (
	"context"
	"errors"
	"time"

	"github.com/feihua/zero-admin/rpc/ums/internal/svc"
	"github.com/feihua/zero-admin/rpc/ums/umsclient"
	"github.com/zeromicro/go-zero/core/logc"

	"github.com/zeromicro/go-zero/core/logx"
)

// AddMemberMessageLogic 创建消息
/*
Author: LiuFeiHua
Date: 2026/4/1
*/
type AddMemberMessageLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAddMemberMessageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddMemberMessageLogic {
	return &AddMemberMessageLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// AddMemberMessage 创建消息
func (l *AddMemberMessageLogic) AddMemberMessage(in *umsclient.AddMemberMessageReq) (*umsclient.AddMemberMessageResp, error) {
	linkType, linkID, relatedOrderID, intentContract, err := normalizeMessageCreatePayload(in)
	if err != nil {
		logc.Errorf(l.ctx, "创建会员消息前归一化意图失败,参数:%+v,异常:%s", in, err.Error())
		return nil, errors.New("创建会员消息失败")
	}

	msg := &memberMessageRecord{
		MemberID:       in.MemberId,
		MessageType:    in.MessageType,
		Title:          in.Title,
		Content:        in.Content,
		ImageURL:       in.ImageUrl,
		LinkType:       linkType,
		LinkID:         linkID,
		RelatedOrderID: relatedOrderID,
		Status:         0,
		CreateTime:     time.Now(),
		PlatformID:     in.PlatformId,
		TenantID:       in.TenantId,
		MerchantID:     in.MerchantId,
		IntentContract: intentContract,
	}

	err = l.svcCtx.DB.WithContext(l.ctx).Create(msg).Error
	if err != nil {
		logc.Errorf(l.ctx, "创建会员消息失败,参数:%+v,异常:%s", in, err.Error())
		return nil, errors.New("创建会员消息失败")
	}

	return &umsclient.AddMemberMessageResp{
		Code:      0,
		Msg:       "创建成功",
		MessageId: msg.ID,
	}, nil
}
