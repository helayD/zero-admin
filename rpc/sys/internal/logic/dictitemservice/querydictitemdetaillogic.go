package dictitemservicelogic

import (
	"context"
	"errors"

	"github.com/bytedance/sonic"
	"github.com/feihua/zero-admin/pkg/time_util"
	logiccommon "github.com/feihua/zero-admin/rpc/sys/internal/logic/common"
	"github.com/zeromicro/go-zero/core/logc"
	"gorm.io/gorm"
	"strconv"

	"github.com/feihua/zero-admin/rpc/sys/internal/svc"
	"github.com/feihua/zero-admin/rpc/sys/sysclient"

	"github.com/zeromicro/go-zero/core/logx"
)

// QueryDictItemDetailLogic QueryDictItemDetail
/*
Author: LiuFeiHua
Date: 2024/5/30 10:27
*/
type QueryDictItemDetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewQueryDictItemDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryDictItemDetailLogic {
	return &QueryDictItemDetailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// QueryDictItemDetail 查询字典数据详情
// 1.判断字典数据是否存在
func (l *QueryDictItemDetailLogic) QueryDictItemDetail(in *sysclient.QueryDictItemDetailReq) (*sysclient.QueryDictItemDetailResp, error) {
	idStr := strconv.FormatInt(in.Id, 10)
	key := l.svcCtx.RedisKey + "dict:item"
	cachedData, _ := l.svcCtx.Redis.HgetCtx(l.ctx, key, idStr)

	var cached sysclient.QueryDictItemDetailResp
	if sonic.Unmarshal([]byte(cachedData), &cached) == nil && cached.Scope != nil {
		return &cached, nil
	}
	var item logiccommon.ScopedDictItem
	err := l.svcCtx.DB.WithContext(l.ctx).
		Table("sys_dict_item").
		Select("id, dict_sort, dict_label, dict_value, dict_type, css_class, list_class, is_default, status, remark, create_by, create_time, update_by, update_time, dict_type_id, platform_id, tenant_id, merchant_id").
		Where("id = ?", in.Id).
		Take(&item).Error

	// 1.判断字典数据是否存在
	switch {
	case errors.Is(err, gorm.ErrRecordNotFound):
		logc.Errorf(l.ctx, "字典数据不存在, 请求参数：%+v, 异常信息: %s", in, err.Error())
		return nil, errors.New("字典数据不存在")
	case err != nil:
		logc.Errorf(l.ctx, "查询字典数据异常, 请求参数：%+v, 异常信息: %s", in, err.Error())
		return nil, errors.New("查询字典数据异常")
	}

	itemScope := logiccommon.DefaultScope(item.PlatformID, item.TenantID, item.MerchantID)
	data := &sysclient.QueryDictItemDetailResp{
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
	}

	value, _ := sonic.Marshal(data)
	filed := strconv.FormatInt(item.ID, 10)
	_ = l.svcCtx.Redis.HsetCtx(l.ctx, key, filed, string(value))
	return data, nil

}
