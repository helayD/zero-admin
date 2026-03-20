package dicttypeservicelogic

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

// AddDictTypeLogic 添加字典信息
/*
Author: LiuFeiHua
Date: 2023/12/18 17:02
*/
type AddDictTypeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAddDictTypeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddDictTypeLogic {
	return &AddDictTypeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// AddDictType 添加字典信息
// 1.查询字典名称是否已存在,如果字典名称已存在,则直接返回
// 2.查询字典类型是否已存在,如果字典类型已存在,则直接返回
// 3.字典不存在时,则直接添加字典
func (l *AddDictTypeLogic) AddDictType(in *sysclient.AddDictTypeReq) (*sysclient.AddDictTypeResp, error) {
	currentScope, err := logiccommon.NormalizeProtoScope(in.Scope)
	if err != nil {
		return nil, err
	}
	if err = logiccommon.ValidateTenantWritable(l.ctx, l.svcCtx.DB, currentScope); err != nil {
		return nil, err
	}

	scopeWhere, scopeArgs := logiccommon.ScopeFilterSQL("", currentScope)
	db := l.svcCtx.DB.WithContext(l.ctx).Table("sys_dict_type")

	// 1.查询字典名称是否已存在,如果字典名称已存在,则直接返回
	var count int64
	err = db.Where(scopeWhere, scopeArgs...).Where("dict_name = ?", in.DictName).Count(&count).Error

	if err != nil {
		logc.Errorf(l.ctx, ".查询字典名称失败,参数：%+v,,异常:%s", in, err.Error())
		return nil, errors.New(fmt.Sprintf("添加字典信息"))
	}

	if count > 0 {
		return nil, errors.New("添加字典类型失败,字典名称已存在")
	}

	// 2.查询字典类型是否已存在,如果字典类型已存在,则直接返回
	err = db.Where(scopeWhere, scopeArgs...).Where("dict_type = ?", in.DictType).Count(&count).Error

	if err != nil {
		logc.Errorf(l.ctx, "查询字典类型失败,参数：%+v,异常:%s", in, err.Error())
		return nil, errors.New(fmt.Sprintf("查询字典信息失败"))
	}

	if count > 0 {
		return nil, errors.New("添加字典类型失败,字典类型已存在")
	}

	dict := &model.SysDictType{
		DictName: in.DictName, // 字典名称
		DictType: in.DictType, // 字典类型
		Status:   in.Status,   // 状态（0：停用，1:正常）
		Remark:   in.Remark,   // 备注
		CreateBy: in.CreateBy, // 创建者
	}
	err = l.svcCtx.DB.WithContext(l.ctx).Transaction(func(tx *gorm.DB) error {
		if err = tx.Create(dict).Error; err != nil {
			return err
		}

		return tx.Table("sys_dict_type").
			Where("id = ?", dict.ID).
			Updates(map[string]interface{}{
				"platform_id": currentScope.PlatformID,
				"tenant_id":   currentScope.TenantID,
				"merchant_id": currentScope.MerchantID,
			}).Error
	})

	if err != nil {
		logc.Errorf(l.ctx, "添加字典信息失败,参数:%+v,异常:%s", dict, err.Error())
		return nil, errors.New("添加字典信息失败")
	}

	return &sysclient.AddDictTypeResp{}, nil
}
