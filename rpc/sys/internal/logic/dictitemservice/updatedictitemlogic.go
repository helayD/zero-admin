package dictitemservicelogic

import (
	"context"
	"errors"
	"fmt"

	logiccommon "github.com/feihua/zero-admin/rpc/sys/internal/logic/common"
	"github.com/feihua/zero-admin/rpc/sys/internal/svc"
	"github.com/feihua/zero-admin/rpc/sys/sysclient"
	"github.com/zeromicro/go-zero/core/logc"
	"gorm.io/gorm"
	"strconv"

	"github.com/zeromicro/go-zero/core/logx"
)

// UpdateDictItemLogic 更新字典数据
/*
Author: LiuFeiHua
Date: 2024/5/28 17:03
*/
type UpdateDictItemLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateDictItemLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateDictItemLogic {
	return &UpdateDictItemLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// UpdateDictItem 更新字典数据
// 1.判断字典数据是否存在
// 2.根据字典类型查询字典是否已存在
// 3.查询字典标签是否已存在,如果字典标签已存在,则直接返回
// 4.查询字典键值是否已存在,如果字典键值已存在,则直接返回
// 5.如果更新字典数据是默认,则修改其他选项为非默认状态
// 6.字典数据存在时,则直接更新字典数据
func (l *UpdateDictItemLogic) UpdateDictItem(in *sysclient.UpdateDictItemReq) (*sysclient.UpdateDictItemResp, error) {
	currentScope, err := logiccommon.NormalizeProtoScope(in.Scope)
	if err != nil {
		return nil, err
	}
	if err = logiccommon.ValidateTenantWritable(l.ctx, l.svcCtx.DB, currentScope); err != nil {
		return nil, err
	}

	// 1.判断字典数据是否存在
	var dictItem logiccommon.ScopedDictItem
	err = l.svcCtx.DB.WithContext(l.ctx).
		Table("sys_dict_item").
		Select("id, dict_sort, dict_label, dict_value, dict_type, css_class, list_class, is_default, status, remark, create_by, create_time, update_by, update_time, dict_type_id, platform_id, tenant_id, merchant_id").
		Where("id = ?", in.Id).
		Take(&dictItem).Error

	// 1.判断字典数据是否存在
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		logc.Errorf(l.ctx, "字典数据不存在, 请求参数：%+v, 异常信息: %s", in, err.Error())
		return nil, errors.New("字典数据不存在")
	case err != nil:
		logc.Errorf(l.ctx, "查询字典数据异常, 请求参数：%+v, 异常信息: %s", in, err.Error())
		return nil, errors.New("查询字典数据异常")
	}
	if err = logiccommon.EnsureScopeMatch(currentScope, dictItem.PlatformID, dictItem.TenantID, dictItem.MerchantID, "不支持跨主体迁移字典项，请在目标主体下新建字典项"); err != nil {
		return nil, err
	}

	// 2.判断字典类型是否存在
	dictType, err := logiccommon.ResolveDictType(l.ctx, l.svcCtx.DB, currentScope, in.DictTypeId, in.DictType)
	if err != nil {
		return nil, err
	}
	scopeWhere, scopeArgs := logiccommon.ScopeFilterSQL("", currentScope)
	db := l.svcCtx.DB.WithContext(l.ctx).Table("sys_dict_item")

	// 3.查询字典标签是否已存在,如果字典标签已存在,则直接返回
	var count int64
	err = db.Where(scopeWhere, scopeArgs...).
		Where("id <> ? AND dict_type_id = ? AND dict_label = ?", in.Id, dictType.ID, in.DictLabel).
		Count(&count).Error

	if err != nil {
		logc.Errorf(l.ctx, "根据字典标签：%s,查询字典数据失败,异常:%s", in.DictLabel, err.Error())
		return nil, errors.New(fmt.Sprintf("更新字典数据失败"))
	}

	if count > 0 {
		logc.Errorf(l.ctx, "更新字典数据失败,字典标签已存在：%+v", in)
		return nil, errors.New(fmt.Sprintf("更新字典数据失败,字典标签：%s,已存在", in.DictLabel))
	}

	// 4.查询字典键值是否已存在,如果字典键值已存在,则直接返回
	err = db.Where(scopeWhere, scopeArgs...).
		Where("id <> ? AND dict_type_id = ? AND dict_value = ?", in.Id, dictType.ID, in.DictValue).
		Count(&count).Error

	if err != nil {
		logc.Errorf(l.ctx, "根据字典键值：%s,查询字典数据失败,异常:%s", in.DictValue, err.Error())
		return nil, errors.New(fmt.Sprintf("更新字典数据失败"))
	}

	if count > 0 {
		logc.Errorf(l.ctx, "更新字典数据失败,字典键值已存在：%+v", in)
		return nil, errors.New(fmt.Sprintf("更新字典数据失败,字典键值：%s,已存在", in.DictValue))
	}

	// 5.如果更新字典数据是默认,则修改其他选项为非默认状态
	if in.IsDefault == "Y" {
		err = db.Where(scopeWhere, scopeArgs...).
			Where("id <> ? AND dict_type_id = ? AND is_default = ?", in.Id, dictType.ID, "Y").
			Updates(map[string]interface{}{
				"is_default": "N",
				"update_by":  in.UpdateBy,
			}).Error
		if err != nil {
			logc.Errorf(l.ctx, "修改字典数据默认状态失败,参数:%+v,异常:%s", in, err.Error())
			return nil, errors.New("更新字典数据失败")
		}
	}

	// 6.字典数据存在时,则直接更新字典数据
	err = l.svcCtx.DB.WithContext(l.ctx).
		Table("sys_dict_item").
		Where("id = ?", in.Id).
		Updates(map[string]interface{}{
			"dict_sort":    in.DictSort,
			"dict_label":   in.DictLabel,
			"dict_value":   in.DictValue,
			"dict_type":    dictType.DictType,
			"css_class":    in.CssClass,
			"list_class":   in.ListClass,
			"is_default":   in.IsDefault,
			"status":       in.Status,
			"remark":       in.Remark,
			"update_by":    in.UpdateBy,
			"dict_type_id": dictType.ID,
			"platform_id":  currentScope.PlatformID,
			"tenant_id":    currentScope.TenantID,
			"merchant_id":  currentScope.MerchantID,
		}).Error

	if err != nil {
		logc.Errorf(l.ctx, "更新字典数据失败,参数:%+v,异常:%s", dictItem, err.Error())
		return nil, errors.New(fmt.Sprintf("更新字典数据失败"))
	}

	key := l.svcCtx.RedisKey + "dict:item"
	filed := strconv.FormatInt(in.Id, 10)
	_, _ = l.svcCtx.Redis.HdelCtx(l.ctx, key, filed)
	return &sysclient.UpdateDictItemResp{}, nil
}
