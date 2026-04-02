package membermessageservicelogic

import (
	"context"
	"errors"
	"time"

	"github.com/feihua/zero-admin/rpc/ums/gen/model"
	"github.com/feihua/zero-admin/rpc/ums/internal/svc"
	"github.com/feihua/zero-admin/rpc/ums/umsclient"
	"github.com/zeromicro/go-zero/core/logc"

	"github.com/zeromicro/go-zero/core/logx"
)

// MarkMessageAsReadLogic 标记单条消息已读
/*
Author: LiuFeiHua
Date: 2026/4/1
*/
type MarkMessageAsReadLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewMarkMessageAsReadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MarkMessageAsReadLogic {
	return &MarkMessageAsReadLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// MarkMessageAsRead 标记单条消息已读
func (l *MarkMessageAsReadLogic) MarkMessageAsRead(in *umsclient.MarkMessageAsReadReq) (*umsclient.MarkMessageAsReadResp, error) {
	// 校验消息归属
	var msg model.UmsMemberMessage
	err := l.svcCtx.DB.WithContext(l.ctx).Where("id = ? AND member_id = ?", in.Id, in.MemberId).First(&msg).Error
	if err != nil {
		logc.Errorf(l.ctx, "标记已读失败,消息不存在,参数:%+v,异常:%s", in, err.Error())
		return &umsclient.MarkMessageAsReadResp{
			Code: 1,
			Msg:  "消息不存在",
		}, nil
	}

	// 如果已经是已读状态，直接返回
	if msg.Status == 1 {
		return &umsclient.MarkMessageAsReadResp{
			Code: 0,
			Msg:  "已是已读状态",
		}, nil
	}

	err = l.svcCtx.DB.WithContext(l.ctx).Model(&model.UmsMemberMessage{}).Where("id = ? AND member_id = ?", in.Id, in.MemberId).Updates(map[string]interface{}{
		"status":    model.MessageStatusRead,
		"read_time": time.Now(),
	}).Error
	if err != nil {
		logc.Errorf(l.ctx, "标记已读失败,参数:%+v,异常:%s", in, err.Error())
		return nil, errors.New("标记已读失败")
	}

	return &umsclient.MarkMessageAsReadResp{
		Code: 0,
		Msg:  "标记成功",
	}, nil
}
