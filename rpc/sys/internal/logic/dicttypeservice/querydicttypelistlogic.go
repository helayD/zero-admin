package dicttypeservicelogic

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

// QueryDictTypeListLogic 查询字典列表
/*
Author: LiuFeiHua
Date: 2023/12/18 17:03
*/
type QueryDictTypeListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewQueryDictTypeListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryDictTypeListLogic {
	return &QueryDictTypeListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// QueryDictTypeList 查询字典表列表
func (l *QueryDictTypeListLogic) QueryDictTypeList(in *sysclient.QueryDictTypeListReq) (*sysclient.QueryDictTypeListResp, error) {
	currentScope, err := logiccommon.NormalizeProtoScope(in.Scope)
	if err != nil {
		return nil, err
	}
	scopeWhere, scopeArgs := logiccommon.ScopeFilterSQL("", currentScope)
	q := l.svcCtx.DB.WithContext(l.ctx).Table("sys_dict_type").Where(scopeWhere, scopeArgs...)
	if len(in.DictName) > 0 {
		q = q.Where("dict_name like ?", "%"+in.DictName+"%")
	}
	if len(in.DictType) > 0 {
		q = q.Where("dict_type like ?", "%"+in.DictType+"%")
	}

	if in.Status != 2 {
		q = q.Where("status = ?", in.Status)
	}

	var count int64
	if err = q.Count(&count).Error; err != nil {
		logc.Errorf(l.ctx, "查询字典列表总数失败,参数：%+v,异常:%s", in, err.Error())
		return nil, errors.New("查询字典列表信息失败")
	}

	var result []logiccommon.ScopedDictType
	err = q.Select("id, dict_name, dict_type, status, remark, create_by, create_time, update_by, update_time, platform_id, tenant_id, merchant_id").
		Order("create_time desc, id desc").
		Offset(int((in.PageNum - 1) * in.PageSize)).
		Limit(int(in.PageSize)).
		Find(&result).Error

	if err != nil {
		logc.Errorf(l.ctx, "查询字典列表信息失败,参数：%+v,异常:%s", in, err.Error())
		return nil, errors.New("查询字典列表信息失败")
	}
	var list = make([]*sysclient.DictTypeListData, 0, len(result))
	for _, dict := range result {
		itemScope := logiccommon.DefaultScope(dict.PlatformID, dict.TenantID, dict.MerchantID)
		list = append(list, &sysclient.DictTypeListData{
			Id:         dict.ID,                                 // 字典id
			DictName:   dict.DictName,                           // 字典名称
			DictType:   dict.DictType,                           // 字典类型
			Status:     dict.Status,                             // 状态（0：停用，1:正常）
			Remark:     dict.Remark,                             // 备注
			CreateBy:   dict.CreateBy,                           // 创建者
			CreateTime: time_util.TimeToStr(dict.CreateTime),    // 创建时间
			UpdateBy:   dict.UpdateBy,                           // 更新者
			UpdateTime: time_util.TimeToString(dict.UpdateTime), // 更新时间
			Scope:      logiccommon.ProtoScope(itemScope),
		})
	}

	return &sysclient.QueryDictTypeListResp{
		Total: count,
		List:  list,
	}, nil
}
