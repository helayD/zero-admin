package membermessageservicelogic

import (
	"context"
	"errors"

	"github.com/feihua/zero-admin/rpc/ums/gen/model"
	"github.com/feihua/zero-admin/rpc/ums/internal/svc"
	"github.com/feihua/zero-admin/rpc/ums/umsclient"
	"github.com/zeromicro/go-zero/core/logc"

	"github.com/zeromicro/go-zero/core/logx"
)

// DeleteMemberMessageLogic 删除消息
/*
Author: LiuFeiHua
Date: 2026/4/1
*/
type DeleteMemberMessageLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteMemberMessageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteMemberMessageLogic {
	return &DeleteMemberMessageLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// DeleteMemberMessage 删除消息
func (l *DeleteMemberMessageLogic) DeleteMemberMessage(in *umsclient.DeleteMemberMessageReq) (*umsclient.DeleteMemberMessageResp, error) {
	// 校验消息归属
	var msg model.UmsMemberMessage
	err := l.svcCtx.DB.WithContext(l.ctx).Where("id = ? AND member_id = ?", in.Id, in.MemberId).First(&msg).Error
	if err != nil {
		logc.Errorf(l.ctx, "删除消息失败,消息不存在或无权访问,参数:%+v,异常:%s", in, err.Error())
		return &umsclient.DeleteMemberMessageResp{
			Code: 1,
			Msg:  "消息不存在或无权访问",
		}, nil
	}

	err = l.svcCtx.DB.WithContext(l.ctx).Where("id = ? AND member_id = ?", in.Id, in.MemberId).Delete(&model.UmsMemberMessage{}).Error
	if err != nil {
		logc.Errorf(l.ctx, "删除消息失败,参数:%+v,异常:%s", in, err.Error())
		return nil, errors.New("删除消息失败")
	}

	return &umsclient.DeleteMemberMessageResp{
		Code: 0,
		Msg:  "删除成功",
	}, nil
}
