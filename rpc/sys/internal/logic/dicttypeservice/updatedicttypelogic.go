package dicttypeservicelogic

import (
	"context"
	"errors"
	"fmt"

	logiccommon "github.com/feihua/zero-admin/rpc/sys/internal/logic/common"
	"github.com/feihua/zero-admin/rpc/sys/internal/svc"
	"github.com/feihua/zero-admin/rpc/sys/sysclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
	"strconv"
)

// UpdateDictTypeLogic 更新字典信息
/*
Author: LiuFeiHua
Date: 2023/12/18 17:03
*/
type UpdateDictTypeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateDictTypeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateDictTypeLogic {
	return &UpdateDictTypeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// UpdateDictType 更新字典表
// 1.根据字典id查询字典是否已存在
// 2.查询字典名称是否已存在,如果字典名称已存在,则直接返回
// 3.查询字典类型是否已存在,如果字典类型已存在,则直接返回
// 4.字典存在时,则直接更新字典
func (l *UpdateDictTypeLogic) UpdateDictType(in *sysclient.UpdateDictTypeReq) (*sysclient.UpdateDictTypeResp, error) {
	currentScope, err := logiccommon.NormalizeProtoScope(in.Scope)
	if err != nil {
		return nil, err
	}
	if err = logiccommon.ValidateTenantWritable(l.ctx, l.svcCtx.DB, currentScope); err != nil {
		return nil, err
	}

	// 1.根据字典id查询字典是否已存在
	var res logiccommon.ScopedDictType
	err = l.svcCtx.DB.WithContext(l.ctx).
		Table("sys_dict_type").
		Select("id, dict_name, dict_type, status, remark, create_by, create_time, update_by, update_time, platform_id, tenant_id, merchant_id").
		Where("id = ?", in.Id).
		Take(&res).Error

	// 1.判断字典类型是否存在
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		logc.Errorf(l.ctx, "字典类型不存在, 请求参数：%+v, 异常信息: %s", in, err.Error())
		return nil, errors.New("字典类型不存在")
	case err != nil:
		logc.Errorf(l.ctx, "查询字典类型异常, 请求参数：%+v, 异常信息: %s", in, err.Error())
		return nil, errors.New("查询字典类型异常")
	}
	if err = logiccommon.EnsureScopeMatch(currentScope, res.PlatformID, res.TenantID, res.MerchantID, "不支持跨主体迁移字典类型，请在目标主体下新建字典类型"); err != nil {
		return nil, err
	}

	scopeWhere, scopeArgs := logiccommon.ScopeFilterSQL("", currentScope)
	db := l.svcCtx.DB.WithContext(l.ctx).Table("sys_dict_type")

	// 2.查询字典名称是否已存在,如果字典名称已存在,则直接返回
	var count int64
	err = db.Where(scopeWhere, scopeArgs...).Where("id <> ? AND dict_name = ?", in.Id, in.DictName).Count(&count).Error

	if err != nil {
		logc.Errorf(l.ctx, ".查询字典名称失败,参数：%+v,,异常:%s", in, err.Error())
		return nil, errors.New(fmt.Sprintf("更新字典信息失败"))
	}

	if count > 0 {
		return nil, errors.New("更新字典类型失败,字典名称已存在")
	}

	// 3.查询字典类型是否已存在,如果字典类型已存在,则直接返回
	err = db.Where(scopeWhere, scopeArgs...).Where("id <> ? AND dict_type = ?", in.Id, in.DictType).Count(&count).Error

	if err != nil {
		logc.Errorf(l.ctx, "查询字典类型失败,参数：%+v,异常:%s", in, err.Error())
		return nil, errors.New(fmt.Sprintf("查询字典信息失败"))
	}

	if count > 0 {
		return nil, errors.New("更新字典类型失败,字典类型已存在")
	}

	// 4.字典存在时,则直接更新字典
	err = l.svcCtx.DB.WithContext(l.ctx).
		Table("sys_dict_type").
		Where("id = ?", in.Id).
		Updates(map[string]interface{}{
			"dict_name":   in.DictName,
			"dict_type":   in.DictType,
			"status":      in.Status,
			"remark":      in.Remark,
			"update_by":   in.UpdateBy,
			"platform_id": currentScope.PlatformID,
			"tenant_id":   currentScope.TenantID,
			"merchant_id": currentScope.MerchantID,
		}).Error

	if err != nil {
		logc.Errorf(l.ctx, "更新字典信息失败,参数:%+v,异常:%s", in, err.Error())
		return nil, errors.New("更新字典信息失败")
	}

	key := l.svcCtx.RedisKey + "dict:type"
	_, _ = l.svcCtx.Redis.HdelCtx(l.ctx, key, strconv.FormatInt(in.Id, 10))
	return &sysclient.UpdateDictTypeResp{}, nil
}
