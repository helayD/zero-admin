package dictitemservicelogic

import (
	"context"
	"errors"

	"github.com/feihua/zero-admin/pkg/time_util"
	logiccommon "github.com/feihua/zero-admin/rpc/sys/internal/logic/common"
	"github.com/zeromicro/go-zero/core/logc"

	"github.com/feihua/zero-admin/rpc/sys/internal/svc"
	"github.com/feihua/zero-admin/rpc/sys/sysclient"

	"github.com/zeromicro/go-zero/core/logx"
)

// QueryDictItemListLogic 查询字典数据列表
/*
Author: LiuFeiHua
Date: 2024/5/28 17:03
*/
type QueryDictItemListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewQueryDictItemListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryDictItemListLogic {
	return &QueryDictItemListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// QueryDictItemList 查询字典数据列表
func (l *QueryDictItemListLogic) QueryDictItemList(in *sysclient.QueryDictItemListReq) (*sysclient.QueryDictItemListResp, error) {
	currentScope, err := logiccommon.NormalizeProtoScope(in.Scope)
	if err != nil {
		return nil, err
	}
	scopeWhere, scopeArgs := logiccommon.ScopeFilterSQL("", currentScope)
	q := l.svcCtx.DB.WithContext(l.ctx).Table("sys_dict_item").Where(scopeWhere, scopeArgs...)

	if len(in.DictType) > 0 {
		q = q.Where("dict_type = ?", in.DictType)
	}
	if in.DictTypeId > 0 {
		q = q.Where("dict_type_id = ?", in.DictTypeId)
	}
	if len(in.DictLabel) > 0 {
		q = q.Where("dict_label like ?", "%"+in.DictLabel+"%")
	}

	if in.Status != 2 {
		q = q.Where("status = ?", in.Status)
	}

	var count int64
	if err = q.Count(&count).Error; err != nil {
		logc.Errorf(l.ctx, "查询字典数据总数失败,参数:%+v,异常:%s", in, err.Error())
		return nil, errors.New("查询字典数据列表失败")
	}

	var result []logiccommon.ScopedDictItem
	err = q.Select("id, dict_sort, dict_label, dict_value, dict_type, css_class, list_class, is_default, status, remark, create_by, create_time, update_by, update_time, dict_type_id, platform_id, tenant_id, merchant_id").
		Order("dict_sort asc, id asc").
		Offset(int((in.PageNum - 1) * in.PageSize)).
		Limit(int(in.PageSize)).
		Find(&result).Error

	if err != nil {
		logc.Errorf(l.ctx, "查询字典数据列表失败,参数:%+v,异常:%s", in, err.Error())
		return nil, errors.New("查询字典数据列表失败")
	}

	var list = make([]*sysclient.DictItemListData, 0, len(result))

	for _, item := range result {
		itemScope := logiccommon.DefaultScope(item.PlatformID, item.TenantID, item.MerchantID)
		list = append(list, &sysclient.DictItemListData{
			Id:         item.ID,                                 // 字典数据id
			DictSort:   item.DictSort,                           // 字典排序
			DictLabel:  item.DictLabel,                          // 字典标签
			DictValue:  item.DictValue,                          // 字典键值
			DictType:   item.DictType,                           // 字典类型
			CssClass:   item.CSSClass,                           // 样式属性（其他样式扩展）
			ListClass:  item.ListClass,                          // 表格回显样式
			IsDefault:  item.IsDefault,                          // 是否默认（Y是 N否）
			Status:     item.Status,                             // 状态（0：停用，1:正常）
			Remark:     item.Remark,                             // 备注
			CreateBy:   item.CreateBy,                           // 创建者
			CreateTime: time_util.TimeToStr(item.CreateTime),    // 创建时间
			UpdateBy:   item.UpdateBy,                           // 更新者
			UpdateTime: time_util.TimeToString(item.UpdateTime), // 更新时间
			DictTypeId: item.DictTypeID,
			Scope:      logiccommon.ProtoScope(itemScope),
		})
	}

	return &sysclient.QueryDictItemListResp{
		Total: count,
		List:  list,
	}, nil
}
