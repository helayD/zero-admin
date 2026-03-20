package postservicelogic

import (
	"context"
	"errors"
	"fmt"

	logiccommon "github.com/feihua/zero-admin/rpc/sys/internal/logic/common"
	"github.com/feihua/zero-admin/rpc/sys/internal/svc"
	"github.com/feihua/zero-admin/rpc/sys/sysclient"
	"github.com/zeromicro/go-zero/core/logc"
	"gorm.io/gorm"

	"github.com/zeromicro/go-zero/core/logx"
)

// UpdatePostLogic 更新岗位信息
/*
Author: LiuFeiHua
Date: 2023/12/18 17:06
*/
type UpdatePostLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdatePostLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdatePostLogic {
	return &UpdatePostLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// UpdatePost 更新岗位信息
// 1.判断岗位信息是否存在
// 2.查询postName是否被占用,如果被占用,则直接返回
// 3.查询根据postCode是否被占用,如果被占用,则直接返回
// 4.岗位存在时,则直接更新岗位
func (l *UpdatePostLogic) UpdatePost(in *sysclient.UpdatePostReq) (*sysclient.UpdatePostResp, error) {
	currentScope, err := logiccommon.NormalizeProtoScope(in.Scope)
	if err != nil {
		return nil, err
	}
	if err = logiccommon.ValidateTenantWritable(l.ctx, l.svcCtx.DB, currentScope); err != nil {
		return nil, err
	}

	// 1.判断岗位信息是否存在
	var post logiccommon.ScopedPost
	err = l.svcCtx.DB.WithContext(l.ctx).
		Table("sys_post").
		Select("id, post_code, post_name, sort, status, remark, create_by, create_time, update_by, update_time, platform_id, tenant_id, merchant_id").
		Where("id = ?", in.Id).
		Take(&post).Error

	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		logc.Errorf(l.ctx, "岗位信息不存在, 请求参数：%+v, 异常信息: %s", in, err.Error())
		return nil, errors.New("岗位信息不存在")
	case err != nil:
		logc.Errorf(l.ctx, "查询岗位信息异常, 请求参数：%+v, 异常信息: %s", in, err.Error())
		return nil, errors.New("查询岗位信息异常")
	}
	if err = logiccommon.EnsureScopeMatch(currentScope, post.PlatformID, post.TenantID, post.MerchantID, "不支持跨主体迁移岗位，请在目标主体下新建岗位"); err != nil {
		return nil, err
	}

	scopeWhere, scopeArgs := logiccommon.ScopeFilterSQL("", currentScope)
	db := l.svcCtx.DB.WithContext(l.ctx).Table("sys_post")

	// 2.查询postName是否被占用,如果被占用,则直接返回
	postName := in.PostName
	var count int64
	err = db.Where(scopeWhere, scopeArgs...).Where("id <> ? AND post_name = ?", in.Id, postName).Count(&count).Error

	if err != nil {
		logc.Errorf(l.ctx, "根据岗位名称：%s,查询岗位信息,异常:%s", postName, err.Error())
		return nil, errors.New(fmt.Sprintf("更新岗位信息失败"))
	}

	if count > 0 {
		return nil, errors.New(fmt.Sprintf("更新岗位信息失败,岗位名称：%s,已存在", postName))
	}

	// 3.查询postCode是否被占用,如果被占用,则直接返回
	postCode := in.PostCode
	err = db.Where(scopeWhere, scopeArgs...).Where("id <> ? AND post_code = ?", in.Id, postCode).Count(&count).Error

	if err != nil {
		logc.Errorf(l.ctx, "根据岗位编码：%s,查询岗位信息,异常:%s", postName, err.Error())
		return nil, errors.New(fmt.Sprintf("更新岗位信息失败"))
	}

	if count > 0 {
		return nil, errors.New(fmt.Sprintf("更新岗位信息失败，岗位编码：%s,已存在", postCode))
	}

	// 4.岗位存在时,则直接更新岗位
	err = l.svcCtx.DB.WithContext(l.ctx).
		Table("sys_post").
		Where("id = ?", in.Id).
		Updates(map[string]interface{}{
			"post_code":   in.PostCode,
			"post_name":   in.PostName,
			"sort":        in.Sort,
			"status":      in.Status,
			"remark":      in.Remark,
			"update_by":   in.UpdateBy,
			"platform_id": currentScope.PlatformID,
			"tenant_id":   currentScope.TenantID,
			"merchant_id": currentScope.MerchantID,
		}).Error

	if err != nil {
		logc.Errorf(l.ctx, "更新岗位信息失败,参数:%+v,异常:%s", in, err.Error())
		return nil, errors.New("更新岗位信息失败")
	}

	return &sysclient.UpdatePostResp{}, nil
}
