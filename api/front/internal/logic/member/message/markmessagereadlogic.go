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

// MarkMessageReadLogic 标记单条消息已读
/*
Author: LiuFeiHua
Date: 2026/4/1
*/
type MarkMessageReadLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewMarkMessageReadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MarkMessageReadLogic {
	return &MarkMessageReadLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// MarkMessageRead 标记单条消息已读
func (l *MarkMessageReadLogic) MarkMessageRead(req *types.MarkMessageReadReq) (resp *types.BaseResp, err error) {
	memberId, err := common.GetMemberId(l.ctx)
	if err != nil {
		return nil, err
	}

	result, err := l.svcCtx.MemberMessageService.MarkMessageAsRead(l.ctx, &umsclient.MarkMessageAsReadReq{
		Id:       req.ID,
		MemberId: memberId,
	})
	if err != nil {
		logc.Errorf(l.ctx, "标记已读失败,参数:%+v,异常:%s", req, err.Error())
		return nil, err
	}

	return &types.BaseResp{
		Code:    result.Code,
		Message: result.Msg,
	}, nil
}
