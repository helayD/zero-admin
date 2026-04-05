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

// MarkAllMessagesReadLogic 标记全部消息已读
/*
Author: LiuFeiHua
Date: 2026/4/1
*/
type MarkAllMessagesReadLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMarkAllMessagesReadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MarkAllMessagesReadLogic {
	return &MarkAllMessagesReadLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// MarkAllMessagesRead 标记全部消息已读
func (l *MarkAllMessagesReadLogic) MarkAllMessagesRead() (resp *types.MarkAllMessagesReadResp, err error) {
	memberId, err := common.GetMemberId(l.ctx)
	if err != nil {
		return nil, err
	}

	result, err := l.svcCtx.MemberMessageService.MarkAllMessagesAsRead(l.ctx, &umsclient.MarkAllMessagesAsReadReq{
		MemberId: memberId,
	})
	if err != nil {
		logc.Errorf(l.ctx, "标记全部已读失败,异常:%s", err.Error())
		return nil, err
	}

	return &types.MarkAllMessagesReadResp{
		Code:         result.Code,
		Message:      result.Msg,
		UpdatedCount: result.UpdatedCount,
	}, nil
}
