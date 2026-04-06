package message

import (
	"context"

	"github.com/feihua/zero-admin/api/front/internal/logic/common"
	"github.com/feihua/zero-admin/api/front/internal/svc"
	"github.com/feihua/zero-admin/api/front/internal/types"
	"github.com/feihua/zero-admin/rpc/ums/umsclient"
	"github.com/zeromicro/go-zero/core/logc"

	"github.com/zeromicro/go-zero/core/logx"
)

// MessageDetailLogic 消息详情
/*
Author: LiuFeiHua
Date: 2026/4/1
*/
type MessageDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMessageDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MessageDetailLogic {
	return &MessageDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// MessageDetail 获取消息详情
func (l *MessageDetailLogic) MessageDetail(req *types.MemberMessageDetailReq) (resp *types.MemberMessageDetailResp, err error) {
	memberId, err := common.GetMemberId(l.ctx)
	if err != nil {
		return nil, err
	}

	result, err := l.svcCtx.MemberMessageService.QueryMemberMessageDetail(l.ctx, &umsclient.QueryMemberMessageDetailReq{
		Id:       req.ID,
		MemberId: memberId,
	})
	if err != nil {
		logc.Errorf(l.ctx, "获取消息详情失败,参数:%+v,异常:%s", req, err.Error())
		return nil, err
	}

	if result.Code != 0 {
		return &types.MemberMessageDetailResp{
			Code:    result.Code,
			Message: result.Msg,
		}, nil
	}

	var data *types.MemberMessageResp
	if result.Data != nil {
		data = &types.MemberMessageResp{
			ID:             result.Data.Id,
			MessageType:    int64(result.Data.MessageType),
			Title:          result.Data.Title,
			Content:        result.Data.Content,
			ImageURL:       result.Data.ImageUrl,
			LinkType:       result.Data.LinkType,
			LinkID:         result.Data.LinkId,
			RelatedOrderID: result.Data.RelatedOrderId,
			Intent:         messageIntentResponseFromRPC(l.ctx, result.Data),
			Status:         int64(result.Data.Status),
			ReadTime:       result.Data.ReadTime,
			CreateTime:     result.Data.CreateTime,
		}
	}

	return &types.MemberMessageDetailResp{
		Code:    result.Code,
		Message: result.Msg,
		Data:    data,
	}, nil
}
