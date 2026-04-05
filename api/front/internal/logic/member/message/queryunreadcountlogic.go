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

// QueryUnreadCountLogic 查询未读消息数量
/*
Author: LiuFeiHua
Date: 2026/4/1
*/
type QueryUnreadCountLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryUnreadCountLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryUnreadCountLogic {
	return &QueryUnreadCountLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// QueryUnreadCount 查询未读消息数量（用于 Flutter Badge）
func (l *QueryUnreadCountLogic) QueryUnreadCount() (resp *types.QueryUnreadCountResp, err error) {
	memberId, err := common.GetMemberId(l.ctx)
	if err != nil {
		return nil, err
	}

	result, err := l.svcCtx.MemberMessageService.QueryUnreadCount(l.ctx, &umsclient.QueryUnreadCountReq{
		MemberId: memberId,
	})
	if err != nil {
		logc.Errorf(l.ctx, "查询未读消息数失败,异常:%s", err.Error())
		return nil, err
	}

	return &types.QueryUnreadCountResp{
		Code:        result.Code,
		Message:     result.Msg,
		UnreadCount: result.UnreadCount,
	}, nil
}
