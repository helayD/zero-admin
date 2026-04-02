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

// DeleteMessageLogic 删除消息
/*
Author: LiuFeiHua
Date: 2026/4/1
*/
type DeleteMessageLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteMessageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteMessageLogic {
	return &DeleteMessageLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// DeleteMessage 删除消息
func (l *DeleteMessageLogic) DeleteMessage(req *types.MemberMessageDetailReq) (resp *types.BaseResp, err error) {
	memberId, err := common.GetMemberId(l.ctx)
	if err != nil {
		return nil, err
	}

	result, err := l.svcCtx.MemberMessageService.DeleteMemberMessage(l.ctx, &umsclient.DeleteMemberMessageReq{
		Id:       req.ID,
		MemberId: memberId,
	})
	if err != nil {
		logc.Errorf(l.ctx, "删除消息失败,参数:%+v,异常:%s", req, err.Error())
		return nil, err
	}

	return &types.BaseResp{
		Code:    result.Code,
		Message: result.Msg,
	}, nil
}
