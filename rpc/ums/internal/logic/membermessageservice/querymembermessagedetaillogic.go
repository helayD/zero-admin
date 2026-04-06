package membermessageservicelogic

import (
	"context"

	"github.com/feihua/zero-admin/rpc/ums/internal/svc"
	"github.com/feihua/zero-admin/rpc/ums/umsclient"
	"github.com/zeromicro/go-zero/core/logc"

	"github.com/zeromicro/go-zero/core/logx"
)

// QueryMemberMessageDetailLogic 查询消息详情
/*
Author: LiuFeiHua
Date: 2026/4/1
*/
type QueryMemberMessageDetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewQueryMemberMessageDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryMemberMessageDetailLogic {
	return &QueryMemberMessageDetailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// QueryMemberMessageDetail 查询消息详情
func (l *QueryMemberMessageDetailLogic) QueryMemberMessageDetail(in *umsclient.QueryMemberMessageDetailReq) (*umsclient.QueryMemberMessageDetailResp, error) {
	var msg memberMessageRecord
	err := l.svcCtx.DB.WithContext(l.ctx).Where("id = ? AND member_id = ?", in.Id, in.MemberId).First(&msg).Error
	if err != nil {
		logc.Errorf(l.ctx, "查询消息详情失败,参数:%+v,异常:%s", in, err.Error())
		return &umsclient.QueryMemberMessageDetailResp{
			Code: 1,
			Msg:  "消息不存在或无权访问",
		}, nil
	}

	item := &umsclient.MemberMessageData{
		Id:             msg.ID,
		MemberId:       msg.MemberID,
		MessageType:    msg.MessageType,
		Title:          msg.Title,
		Content:        msg.Content,
		ImageUrl:       msg.ImageURL,
		LinkType:       msg.LinkType,
		LinkId:         msg.LinkID,
		RelatedOrderId: msg.RelatedOrderID,
		Status:         msg.Status,
		CreateTime:     msg.CreateTime.Format("2006-01-02 15:04:05"),
		Intent:         resolvePersistedIntent(&msg),
	}
	if msg.ReadTime != nil {
		item.ReadTime = msg.ReadTime.Format("2006-01-02 15:04:05")
	}

	return &umsclient.QueryMemberMessageDetailResp{
		Code: 0,
		Msg:  "查询成功",
		Data: item,
	}, nil
}
