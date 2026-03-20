package postservicelogic

import (
	"context"
	"errors"

	"github.com/feihua/zero-admin/pkg/time_util"
	logiccommon "github.com/feihua/zero-admin/rpc/sys/internal/logic/common"
	"github.com/feihua/zero-admin/rpc/sys/sysclient"
	"github.com/zeromicro/go-zero/core/logc"

	"github.com/feihua/zero-admin/rpc/sys/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

// QueryPostListLogic 查询岗位列表
/*
Author: LiuFeiHua
Date: 2023/12/18 17:05
*/
type QueryPostListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewQueryPostListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryPostListLogic {
	return &QueryPostListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// QueryPostList 查询岗位列表
func (l *QueryPostListLogic) QueryPostList(in *sysclient.QueryPostListReq) (*sysclient.QueryPostListResp, error) {
	currentScope, err := logiccommon.NormalizeProtoScope(in.Scope)
	if err != nil {
		return nil, err
	}
	scopeWhere, scopeArgs := logiccommon.ScopeFilterSQL("", currentScope)
	q := l.svcCtx.DB.WithContext(l.ctx).Table("sys_post").Where(scopeWhere, scopeArgs...)
	if len(in.PostCode) > 0 {
		q = q.Where("post_code like ?", "%"+in.PostCode+"%")
	}
	if len(in.PostName) > 0 {
		q = q.Where("post_name like ?", "%"+in.PostName+"%")
	}

	if in.Status != 2 {
		q = q.Where("status = ?", in.Status)
	}

	var count int64
	if err = q.Count(&count).Error; err != nil {
		logc.Errorf(l.ctx, "查询岗位列表总数失败,参数：%+v,异常:%s", in, err.Error())
		return nil, errors.New("查询岗位列表信息失败")
	}

	var result []logiccommon.ScopedPost
	err = q.Select("id, post_code, post_name, sort, status, remark, create_by, create_time, update_by, update_time, platform_id, tenant_id, merchant_id").
		Order("sort asc, id asc").
		Offset(int((in.PageNum - 1) * in.PageSize)).
		Limit(int(in.PageSize)).
		Find(&result).Error

	if err != nil {
		logc.Errorf(l.ctx, "查询岗位列表信息失败,参数：%+v,异常:%s", in, err.Error())
		return nil, errors.New("查询岗位列表信息失败")
	}

	var list = make([]*sysclient.PostListData, 0, len(result))
	for _, post := range result {
		itemScope := logiccommon.DefaultScope(post.PlatformID, post.TenantID, post.MerchantID)
		list = append(list, &sysclient.PostListData{
			Id:         post.ID,                                 // 岗位id
			PostCode:   post.PostCode,                           // 岗位编码
			PostName:   post.PostName,                           // 岗位名称
			Sort:       post.Sort,                               // 显示顺序
			Status:     post.Status,                             // 岗位状态（0：停用，1:正常）
			Remark:     post.Remark,                             // 备注
			CreateBy:   post.CreateBy,                           // 创建者
			CreateTime: time_util.TimeToStr(post.CreateTime),    // 创建时间
			UpdateBy:   post.UpdateBy,                           // 更新者
			UpdateTime: time_util.TimeToString(post.UpdateTime), // 更新时间
			Scope:      logiccommon.ProtoScope(itemScope),
		})
	}

	return &sysclient.QueryPostListResp{
		Total: count,
		List:  list,
	}, nil
}
