package membermessageservicelogic

import (
	"context"

	"github.com/feihua/zero-admin/rpc/ums/gen/model"
	"github.com/feihua/zero-admin/rpc/ums/internal/svc"
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
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewQueryUnreadCountLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryUnreadCountLogic {
	return &QueryUnreadCountLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// QueryUnreadCount 查询未读消息数量
func (l *QueryUnreadCountLogic) QueryUnreadCount(in *umsclient.QueryUnreadCountReq) (*umsclient.QueryUnreadCountResp, error) {
	var count int64
	err := l.svcCtx.DB.WithContext(l.ctx).Model(&model.UmsMemberMessage{}).
		Where("member_id = ? AND status = ?", in.MemberId, model.MessageStatusUnread).
		Count(&count).Error
	if err != nil {
		logc.Errorf(l.ctx, "查询未读消息数失败,参数:%+v,异常:%s", in, err.Error())
		return nil, err
	}

	return &umsclient.QueryUnreadCountResp{
		Code:        0,
		Msg:         "查询成功",
		UnreadCount: count,
	}, nil
}
