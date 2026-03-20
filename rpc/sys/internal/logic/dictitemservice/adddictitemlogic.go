package dictitemservicelogic

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
	"time"
)

// AddDictItemLogic 添加字典数据
/*
Author: LiuFeiHua
Date: 2024/5/28 17:03
*/
type AddDictItemLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAddDictItemLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddDictItemLogic {
	return &AddDictItemLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// AddDictItem 添加字典数据
// 1.根据字典类型查询字典是否已存在
// 2.查询字典标签是否已存在,如果字典标签已存在,则直接返回
// 3.查询字典键值是否已存在,如果字典键值已存在,则直接返回
// 4.如果新增字典数据是默认,则修改其他选项为非默认状态
// 5.字典数据不存在时,则直接添加字典数据
func (l *AddDictItemLogic) AddDictItem(in *sysclient.AddDictItemReq) (*sysclient.AddDictItemResp, error) {
	currentScope, err := logiccommon.NormalizeProtoScope(in.Scope)
	if err != nil {
		return nil, err
	}
	if err = logiccommon.ValidateTenantWritable(l.ctx, l.svcCtx.DB, currentScope); err != nil {
		return nil, err
	}

	dictType, err := logiccommon.ResolveDictType(l.ctx, l.svcCtx.DB, currentScope, in.DictTypeId, in.DictType)
	if err != nil {
		return nil, err
	}
	scopeWhere, scopeArgs := logiccommon.ScopeFilterSQL("", currentScope)
	db := l.svcCtx.DB.WithContext(l.ctx).Table("sys_dict_item")

	// 2.查询字典标签是否已存在,如果字典标签已存在,则直接返回
	var count int64
	err = db.Where(scopeWhere, scopeArgs...).
		Where("dict_type_id = ? AND dict_label = ?", dictType.ID, in.DictLabel).
		Count(&count).Error

	if err != nil {
		logc.Errorf(l.ctx, "根据字典标签：%s,查询字典数据信息失败,异常:%s", in.DictLabel, err.Error())
		return nil, errors.New(fmt.Sprintf("新增字典数据失败"))
	}

	if count > 0 {
		return nil, errors.New(fmt.Sprintf("新增字典数据失败,字典标签：%s,已存在", in.DictLabel))
	}

	// 3.查询字典键值是否已存在,如果字典键值已存在,则直接返回
	err = db.Where(scopeWhere, scopeArgs...).
		Where("dict_type_id = ? AND dict_value = ?", dictType.ID, in.DictValue).
		Count(&count).Error

	if err != nil {
		logc.Errorf(l.ctx, "根据字典键值：%s,查询字典数据信息失败,异常:%s", in.DictValue, err.Error())
		return nil, errors.New(fmt.Sprintf("新增字典数据失败"))
	}

	if count > 0 {
		return nil, errors.New(fmt.Sprintf("新增字典数据失败,字典键值：%s,已存在", in.DictValue))
	}

	// 4.如果新增字典数据是默认,则修改其他选项为非默认状态
	if in.IsDefault == "Y" {
		now := time.Now()
		err = db.Where(scopeWhere, scopeArgs...).
			Where("dict_type_id = ? AND is_default = ?", dictType.ID, "Y").
			Updates(map[string]interface{}{
				"is_default":  "N",
				"update_by":   in.CreateBy,
				"update_time": &now,
			}).Error
		if err != nil {
			logc.Errorf(l.ctx, "修改字典数据默认状态失败,参数:%+v,异常:%s", in, err.Error())
			return nil, errors.New("新增字典数据失败")
		}
	}

	// 5.字典数据不存在时,则直接添加字典数据
	dictItem := &model.SysDictItem{
		DictSort:  in.DictSort,       // 字典排序
		DictLabel: in.DictLabel,      // 字典标签
		DictValue: in.DictValue,      // 字典键值
		DictType:  dictType.DictType, // 字典类型
		CSSClass:  in.CssClass,       // 样式属性（其他样式扩展）
		ListClass: in.ListClass,      // 表格回显样式
		IsDefault: in.IsDefault,      // 是否默认（Y是 N否）
		Status:    in.Status,         // 状态（0：停用，1:正常）
		Remark:    in.Remark,         // 备注
		CreateBy:  in.CreateBy,       // 创建者
	}

	err = l.svcCtx.DB.WithContext(l.ctx).Transaction(func(tx *gorm.DB) error {
		if err = tx.Create(dictItem).Error; err != nil {
			return err
		}

		return tx.Table("sys_dict_item").
			Where("id = ?", dictItem.ID).
			Updates(map[string]interface{}{
				"dict_type_id": dictType.ID,
				"platform_id":  currentScope.PlatformID,
				"tenant_id":    currentScope.TenantID,
				"merchant_id":  currentScope.MerchantID,
			}).Error
	})

	if err != nil {
		logc.Errorf(l.ctx, "添加字典数据信息失败,参数:%+v,异常:%s", dictItem, err.Error())
		return nil, errors.New("添加字典数据信息失败")
	}

	return &sysclient.AddDictItemResp{}, nil
}
