package postservicelogic

import (
	"context"
	"errors"
	"fmt"

	"github.com/feihua/zero-admin/rpc/sys/gen/model"
	logiccommon "github.com/feihua/zero-admin/rpc/sys/internal/logic/common"
	"github.com/feihua/zero-admin/rpc/sys/internal/svc"
	"github.com/feihua/zero-admin/rpc/sys/sysclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

// AddPostLogic 添加岗位
/*
Author: LiuFeiHua
Date: 2023/12/18 17:04
*/
type AddPostLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAddPostLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddPostLogic {
	return &AddPostLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// AddPost 添加岗位
// 1.根据postName查询岗位是否已存在,如果岗位已存在,则直接返回
// 2.根据postCode查询岗位是否已存在,如果岗位已存在,则直接返回
// 3.岗位不存在时,则直接添加岗位
func (l *AddPostLogic) AddPost(in *sysclient.AddPostReq) (*sysclient.AddPostResp, error) {
	currentScope, err := logiccommon.NormalizeProtoScope(in.Scope)
	if err != nil {
		return nil, err
	}
	if err = logiccommon.ValidateTenantWritable(l.ctx, l.svcCtx.DB, currentScope); err != nil {
		return nil, err
	}

	scopeWhere, scopeArgs := logiccommon.ScopeFilterSQL("", currentScope)
	db := l.svcCtx.DB.WithContext(l.ctx).Table("sys_post")

	// 1.根据postName查询岗位是否已存在,如果岗位已存在,则直接返回
	postName := in.PostName
	var count int64
	err = db.Where(scopeWhere, scopeArgs...).Where("post_name = ?", postName).Count(&count).Error

	if err != nil {
		logc.Errorf(l.ctx, "根据岗位名称：%s,查询岗位信息,异常:%s", postName, err.Error())
		return nil, errors.New(fmt.Sprintf("添加岗位信息失败"))
	}

	if count > 0 {
		return nil, errors.New(fmt.Sprintf("添加岗位信息失败,岗位名称：%s,已存在", postName))
	}

	// 2.根据postCode查询岗位是否已存在,如果岗位已存在,则直接返回
	postCode := in.PostCode
	err = db.Where(scopeWhere, scopeArgs...).Where("post_code = ?", postCode).Count(&count).Error

	if err != nil {
		logc.Errorf(l.ctx, "根据岗位编码：%s,查询岗位信息,异常:%s", postName, err.Error())
		return nil, errors.New(fmt.Sprintf("添加岗位信息失败"))
	}

	if count > 0 {
		return nil, errors.New(fmt.Sprintf("添加岗位信息失败，岗位编码：%s,已存在", postCode))
	}

	// 3.岗位不存在时,则直接添加岗位
	job := &model.SysPost{
		PostCode: in.PostCode, // 岗位编码
		PostName: in.PostName, // 岗位名称
		Sort:     in.Sort,     // 显示顺序
		Status:   in.Status,   // 岗位状态（0：停用，1:正常）
		Remark:   in.Remark,   // 备注
		CreateBy: in.CreateBy, // 创建者
	}

	err = l.svcCtx.DB.WithContext(l.ctx).Transaction(func(tx *gorm.DB) error {
		if err = tx.Create(job).Error; err != nil {
			return err
		}

		return tx.Table("sys_post").
			Where("id = ?", job.ID).
			Updates(map[string]interface{}{
				"platform_id": currentScope.PlatformID,
				"tenant_id":   currentScope.TenantID,
				"merchant_id": currentScope.MerchantID,
			}).Error
	})

	if err != nil {
		logc.Errorf(l.ctx, "添加岗位管理失败,参数:%+v,异常:%s", in, err.Error())
		return nil, errors.New("添加岗位信息失败")
	}

	return &sysclient.AddPostResp{}, nil
}
