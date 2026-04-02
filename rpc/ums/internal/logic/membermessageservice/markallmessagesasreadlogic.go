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

// MarkAllMessagesAsReadLogic 标记全部消息已读
/*
Author: LiuFeiHua
Date: 2026/4/1
*/
type MarkAllMessagesAsReadLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewMarkAllMessagesAsReadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *MarkAllMessagesAsReadLogic {
	return &MarkAllMessagesAsReadLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// MarkAllMessagesAsRead 标记全部消息已读
func (l *MarkAllMessagesAsReadLogic) MarkAllMessagesAsRead(in *umsclient.MarkAllMessagesAsReadReq) (*umsclient.MarkAllMessagesAsReadResp, error) {
	result := l.svcCtx.DB.WithContext(l.ctx).Model(&model.UmsMemberMessage{}).
		Where("member_id = ? AND status = ?", in.MemberId, model.MessageStatusUnread).
		Updates(map[string]interface{}{
			"status":    model.MessageStatusRead,
			"read_time": time.Now(),
		})

	if result.Error != nil {
		logc.Errorf(l.ctx, "标记全部已读失败,参数:%+v,异常:%s", in, result.Error.Error())
		return nil, errors.New("标记全部已读失败")
	}

	return &umsclient.MarkAllMessagesAsReadResp{
		Code:         0,
		Msg:          "标记成功",
		UpdatedCount: result.RowsAffected,
	}, nil
}
