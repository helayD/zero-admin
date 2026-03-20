package userservicelogic

import (
	"context"
	"github.com/feihua/zero-admin/pkg/time_util"
	logiccommon "github.com/feihua/zero-admin/rpc/sys/internal/logic/common"
	"github.com/feihua/zero-admin/rpc/sys/internal/svc"
	"github.com/feihua/zero-admin/rpc/sys/sysclient"
	"github.com/zeromicro/go-zero/core/logx"
)

// QueryDeptAndPostListLogic 查询所有部门和岗位
/*
Author: LiuFeiHua
Date: 2024/5/30 9:22
*/
type QueryDeptAndPostListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewQueryDeptAndPostListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryDeptAndPostListLogic {
	return &QueryDeptAndPostListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// QueryDeptAndPostList 查询所有部门和岗位
// 1.查询所有部门
// 2.查询所有岗位
func (l *QueryDeptAndPostListLogic) QueryDeptAndPostList(in *sysclient.QueryDeptAndPostListReq) (*sysclient.QueryDeptAndPostListResp, error) {
	currentScope, err := logiccommon.NormalizeProtoScope(in.Scope)
	if err != nil {
		return nil, err
	}
	scopeWhere, scopeArgs := logiccommon.ScopeFilterSQL("", currentScope)

	// 1.查询所有部门
	var deptList []logiccommon.ScopedDept
	_ = l.svcCtx.DB.WithContext(l.ctx).
		Table("sys_dept").
		Select("id, parent_id, ancestors, dept_name, sort, leader, phone, email, status, del_flag, remark, create_by, create_time, update_by, update_time, platform_id, tenant_id, merchant_id").
		Where(scopeWhere, scopeArgs...).
		Find(&deptList).Error
	var deptListData = make([]*sysclient.DeptData, 0, len(deptList))

	for _, dept := range deptList {
		deptListData = append(deptListData, &sysclient.DeptData{
			Id:         dept.ID,                                 // 部门id
			ParentId:   dept.ParentID,                           // 上级部门id
			Ancestors:  dept.Ancestors,                          // 祖级列表
			DeptName:   dept.DeptName,                           // 部门名称
			Sort:       dept.Sort,                               // 显示顺序
			Leader:     dept.Leader,                             // 负责人
			Phone:      dept.Phone,                              // 联系电话
			Email:      dept.Email,                              // 邮箱
			Status:     dept.Status,                             // 部门状态（0：停用，1:正常）
			DelFlag:    dept.DelFlag,                            // 删除标志（0代表删除 1代表存在）
			Remark:     dept.Remark,                             // 备注信息
			CreateBy:   dept.CreateBy,                           // 创建者
			CreateTime: time_util.TimeToStr(dept.CreateTime),    // 创建时间
			UpdateBy:   dept.UpdateBy,                           // 更新者
			UpdateTime: time_util.TimeToString(dept.UpdateTime), // 更新时间
			Scope:      logiccommon.ProtoScope(logiccommon.DefaultScope(dept.PlatformID, dept.TenantID, dept.MerchantID)),
		})
	}

	// 2.查询所有岗位
	var postList []logiccommon.ScopedPost
	_ = l.svcCtx.DB.WithContext(l.ctx).
		Table("sys_post").
		Select("id, post_code, post_name, sort, status, remark, create_by, create_time, update_by, update_time, platform_id, tenant_id, merchant_id").
		Where(scopeWhere, scopeArgs...).
		Find(&postList).Error
	var postListData = make([]*sysclient.PostData, 0, len(postList))

	for _, post := range postList {
		postListData = append(postListData, &sysclient.PostData{
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
			Scope:      logiccommon.ProtoScope(logiccommon.DefaultScope(post.PlatformID, post.TenantID, post.MerchantID)),
		})
	}

	data := &sysclient.QueryDeptAndPostListResp{
		DeptListData: deptListData,
		PostListData: postListData,
	}

	return data, nil
}
