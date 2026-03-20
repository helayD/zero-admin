package noticeservicelogic

import (
	"context"
	"errors"

	"github.com/feihua/zero-admin/pkg/time_util"
	logiccommon "github.com/feihua/zero-admin/rpc/sys/internal/logic/common"
	"github.com/feihua/zero-admin/rpc/sys/internal/svc"
	"github.com/feihua/zero-admin/rpc/sys/sysclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
)

// QueryNoticeListLogic 查询通知公告列表
/*
Author: 刘飞华
Date: 2025/10/27 15:51:14
*/
type QueryNoticeListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewQueryNoticeListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryNoticeListLogic {
	return &QueryNoticeListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// QueryNoticeList 查询通知公告列表
func (l *QueryNoticeListLogic) QueryNoticeList(in *sysclient.QueryNoticeListReq) (*sysclient.QueryNoticeListResp, error) {
	currentScope, err := logiccommon.NormalizeProtoScope(in.Scope)
	if err != nil {
		return nil, err
	}
	scopeWhere, scopeArgs := logiccommon.ScopeFilterSQL("", currentScope)
	q := l.svcCtx.DB.WithContext(l.ctx).Table("sys_notice").Where(scopeWhere, scopeArgs...)
	if len(in.NoticeTitle) > 0 {
		q = q.Where("notice_title like ?", "%"+in.NoticeTitle+"%")
	}
	if in.NoticeType != 0 {
		q = q.Where("notice_type = ?", in.NoticeType)
	}

	if in.Status != 2 {
		q = q.Where("status = ?", in.Status)
	}

	var count int64
	if err = q.Count(&count).Error; err != nil {
		logc.Errorf(l.ctx, "查询通知公告总数失败,参数:%+v,异常:%s", in, err.Error())
		return nil, errors.New("查询通知公告列表失败")
	}

	var result []logiccommon.ScopedNotice
	err = q.Select("id, notice_title, notice_type, notice_content, status, remark, create_by, create_time, update_by, update_time, platform_id, tenant_id, merchant_id").
		Order("create_time desc, id desc").
		Offset(int((in.PageNum - 1) * in.PageSize)).
		Limit(int(in.PageSize)).
		Find(&result).Error

	if err != nil {
		logc.Errorf(l.ctx, "查询通知公告列表失败,参数:%+v,异常:%s", in, err.Error())
		return nil, errors.New("查询通知公告列表失败")
	}

	var list []*sysclient.QueryNoticeListData

	for _, item := range result {
		itemScope := logiccommon.DefaultScope(item.PlatformID, item.TenantID, item.MerchantID)
		list = append(list, &sysclient.QueryNoticeListData{
			Id:            item.ID,                                 // 公告ID
			NoticeTitle:   item.NoticeTitle,                        // 公告标题
			NoticeType:    item.NoticeType,                         // 公告类型（1:通知,2:公告）
			NoticeContent: item.NoticeContent,                      // 公告内容
			Status:        item.Status,                             // 公告状态（0:关闭,1:正常 ）
			Remark:        item.Remark,                             // 备注
			CreateBy:      item.CreateBy,                           // 创建者
			CreateTime:    time_util.TimeToStr(item.CreateTime),    // 创建时间
			UpdateBy:      item.UpdateBy,                           // 更新者
			UpdateTime:    time_util.TimeToString(item.UpdateTime), // 更新时间
			Scope:         logiccommon.ProtoScope(itemScope),
		})
	}

	return &sysclient.QueryNoticeListResp{
		Total: count,
		List:  list,
	}, nil
}
