package noticeservicelogic

import (
	"context"
	"errors"
	"fmt"

	"github.com/feihua/zero-admin/pkg/errorx"
	logiccommon "github.com/feihua/zero-admin/rpc/sys/internal/logic/common"
	"github.com/feihua/zero-admin/rpc/sys/internal/svc"
	"github.com/feihua/zero-admin/rpc/sys/sysclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

// UpdateNoticeLogic 更新通知公告
/*
Author: 刘飞华
Date: 2025/10/27 15:49:44
*/
type UpdateNoticeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateNoticeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateNoticeLogic {
	return &UpdateNoticeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// UpdateNotice 更新通知公告
func (l *UpdateNoticeLogic) UpdateNotice(in *sysclient.UpdateNoticeReq) (*sysclient.UpdateNoticeResp, error) {
	currentScope, err := logiccommon.NormalizeProtoScope(in.Scope)
	if err != nil {
		return nil, err
	}
	if err = logiccommon.ValidateTenantWritable(l.ctx, l.svcCtx.DB, currentScope); err != nil {
		return nil, err
	}

	// 1.根据通知公告id查询通知公告是否已存在
	var sysNotice logiccommon.ScopedNotice
	err = l.svcCtx.DB.WithContext(l.ctx).
		Table("sys_notice").
		Select("id, notice_title, notice_type, notice_content, status, remark, create_by, create_time, update_by, update_time, platform_id, tenant_id, merchant_id").
		Where("id = ?", in.Id).
		Take(&sysNotice).Error

	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		logc.Errorf(l.ctx, "通知公告不存在, 请求参数：%+v, 异常信息: %s", in, err.Error())
		return nil, errorx.NewDefaultError("通知公告不存在")
	case err != nil:
		logc.Errorf(l.ctx, "查询通知公告异常, 请求参数：%+v, 异常信息: %s", in, err.Error())
		return nil, errorx.NewDefaultError("查询通知公告异常")
	}
	if err = logiccommon.EnsureScopeMatch(currentScope, sysNotice.PlatformID, sysNotice.TenantID, sysNotice.MerchantID, "不支持跨主体迁移通知，请在目标主体下新建通知"); err != nil {
		return nil, err
	}

	scopeWhere, scopeArgs := logiccommon.ScopeFilterSQL("", currentScope)
	var count int64
	err = l.svcCtx.DB.WithContext(l.ctx).
		Table("sys_notice").
		Where(scopeWhere, scopeArgs...).
		Where("id <> ? AND notice_title = ?", in.Id, in.NoticeTitle).
		Count(&count).Error
	if err != nil {
		logc.Errorf(l.ctx, "根据公告标题：%s,查询公告信息,异常:%s", in.NoticeTitle, err.Error())
		return nil, errors.New(fmt.Sprintf("添加通知公告失败"))
	}
	if count > 0 {
		return nil, errors.New(fmt.Sprintf("添加通知公告失败,公告标题：%s,已存在", in.NoticeTitle))
	}

	// 2.通知公告存在时,则直接更新通知公告
	err = l.svcCtx.DB.WithContext(l.ctx).
		Table("sys_notice").
		Where("id = ?", in.Id).
		Updates(map[string]interface{}{
			"notice_title":   in.NoticeTitle,
			"notice_type":    in.NoticeType,
			"notice_content": in.NoticeContent,
			"status":         in.Status,
			"remark":         in.Remark,
			"update_by":      in.UpdateBy,
			"platform_id":    currentScope.PlatformID,
			"tenant_id":      currentScope.TenantID,
			"merchant_id":    currentScope.MerchantID,
		}).Error

	if err != nil {
		logc.Errorf(l.ctx, "更新通知公告失败,参数:%+v,异常:%s", in, err.Error())
		return nil, errors.New("更新通知公告失败")
	}

	return &sysclient.UpdateNoticeResp{}, nil
}
