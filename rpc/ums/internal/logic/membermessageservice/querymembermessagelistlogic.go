package membermessageservicelogic

import (
	"context"
	"errors"

	"github.com/feihua/zero-admin/rpc/ums/internal/svc"
	"github.com/feihua/zero-admin/rpc/ums/umsclient"
	"github.com/zeromicro/go-zero/core/logc"

	"github.com/zeromicro/go-zero/core/logx"
)

// QueryMemberMessageListLogic 查询消息列表
/*
Author: LiuFeiHua
Date: 2026/4/1
*/
type QueryMemberMessageListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewQueryMemberMessageListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryMemberMessageListLogic {
	return &QueryMemberMessageListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// QueryMemberMessageList 查询消息列表
func (l *QueryMemberMessageListLogic) QueryMemberMessageList(in *umsclient.QueryMemberMessageListReq) (*umsclient.QueryMemberMessageListResp, error) {
	var messages []*memberMessageRecord
	var total int64
	var err error

	queryBuilder := l.svcCtx.DB.WithContext(l.ctx).Model(&memberMessageRecord{}).Where("member_id = ?", in.MemberId)

	// 按消息类型筛选
	if in.MessageType > 0 {
		queryBuilder = queryBuilder.Where("message_type = ?", in.MessageType)
	}

	// 按状态筛选（仅当明确传入状态参数时才筛选）
	if in.Status != nil {
		queryBuilder = queryBuilder.Where("status = ?", *in.Status)
	}

	// 统计总数
	if err = queryBuilder.Count(&total).Error; err != nil {
		logc.Errorf(l.ctx, "查询消息列表总数失败,参数:%+v,异常:%s", in, err.Error())
		return nil, errors.New("查询消息列表失败")
	}

	// 分页查询
	offset := (in.PageNum - 1) * in.PageSize
	if err = queryBuilder.Order("create_time DESC").Offset(int(offset)).Limit(int(in.PageSize)).Find(&messages).Error; err != nil {
		logc.Errorf(l.ctx, "查询消息列表失败,参数:%+v,异常:%s", in, err.Error())
		return nil, errors.New("查询消息列表失败")
	}

	// 转换结果
	list := make([]*umsclient.MemberMessageData, 0, len(messages))
	for _, msg := range messages {
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
			Intent:         resolvePersistedIntent(msg),
		}
		if msg.ReadTime != nil {
			item.ReadTime = msg.ReadTime.Format("2006-01-02 15:04:05")
		}
		list = append(list, item)
	}

	return &umsclient.QueryMemberMessageListResp{
		Total:    total,
		PageNum:  in.PageNum,
		PageSize: in.PageSize,
		List:     list,
	}, nil
}
