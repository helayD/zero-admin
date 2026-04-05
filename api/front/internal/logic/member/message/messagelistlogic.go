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

// MessageListLogic 消息列表
/*
Author: LiuFeiHua
Date: 2026/4/1
*/
type MessageListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMessageListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MessageListLogic {
	return &MessageListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// MessageList 获取消息列表
func (l *MessageListLogic) MessageList(req *types.QueryMemberMessageReq) (resp *types.QueryMemberMessageListResp, err error) {
	memberId, err := common.GetMemberId(l.ctx)
	if err != nil {
		return nil, err
	}

	// 处理状态筛选参数（仅当明确传入时传递）
	var status *int32
	if req.Status >= 0 {
		s := int32(req.Status)
		status = &s
	}

	result, err := l.svcCtx.MemberMessageService.QueryMemberMessageList(l.ctx, &umsclient.QueryMemberMessageListReq{
		MemberId:    memberId,
		MessageType: int32(req.MessageType),
		Status:      status,
		PageNum:     req.PageNum,
		PageSize:    req.PageSize,
	})
	if err != nil {
		logc.Errorf(l.ctx, "获取消息列表失败,参数:%+v,异常:%s", req, err.Error())
		return nil, err
	}

	list := make([]types.MemberMessageResp, 0, len(result.List))
	for _, item := range result.List {
		list = append(list, types.MemberMessageResp{
			ID:             item.Id,
			MessageType:    int64(item.MessageType),
			Title:          item.Title,
			Content:        item.Content,
			ImageURL:       item.ImageUrl,
			LinkType:       item.LinkType,
			LinkID:         item.LinkId,
			RelatedOrderID: item.RelatedOrderId,
			Status:         int64(item.Status),
			ReadTime:       item.ReadTime,
			CreateTime:     item.CreateTime,
		})
	}

	return &types.QueryMemberMessageListResp{
		Code:    0,
		Message: "查询成功",
		Data:    list,
		Total:   result.Total,
		PageNum:  result.PageNum,
		PageSize: result.PageSize,
	}, nil
}
