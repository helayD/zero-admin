package role

import (
	"context"

	"github.com/feihua/zero-admin/api/admin/internal/common/errorx"
	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/feihua/zero-admin/rpc/sys/sysclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/status"
)

// QueryRoleListLogic 查询角色列表
/*
Author: LiuFeiHua
Date: 2023/12/18 15:39
*/
type QueryRoleListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryRoleListLogic(ctx context.Context, svcCtx *svc.ServiceContext) QueryRoleListLogic {
	return QueryRoleListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// QueryRoleList 查询角色列表
func (l *QueryRoleListLogic) QueryRoleList(req *types.QueryRoleListReq) (*types.QueryRoleListResp, error) {
	result, err := l.svcCtx.RoleService.QueryRoleList(l.ctx, &sysclient.QueryRoleListReq{
		PageNum:    req.Current,
		PageSize:   req.PageSize,
		RoleName:   req.RoleName,   // 名称
		RoleKey:    req.RoleKey,    // 角色权限字符串
		DataScope:  req.DataScope,  // 数据范围
		Status:     req.Status,     // 状态(1:正常，0:禁用)
		ScopeType:  req.ScopeType,  // 作用域类型过滤
		TenantId:   req.TenantId,   // 租户ID过滤
		MerchantId: req.MerchantId, // 商户ID过滤
	})

	if err != nil {
		logc.Errorf(l.ctx, "查询角色列表,参数: %+v,异常:%s", req, err.Error())
		s, _ := status.FromError(err)
		return nil, errorx.NewDefaultError(s.Message())
	}

	var list []*types.QueryRoleListData

	for _, detail := range result.List {
		list = append(list, &types.QueryRoleListData{
			Id:         detail.Id,         // 角色id
			RoleName:   detail.RoleName,   // 名称
			RoleKey:    detail.RoleKey,    // 角色权限字符串
			DataScope:  detail.DataScope,  // 数据范围
			Status:     detail.Status,     // 状态(1:正常，0:禁用)
			Remark:     detail.Remark,     // 备注
			DelFlag:    detail.DelFlag,    // 删除标志
			CreateBy:   detail.CreateBy,   // 创建者
			CreateTime: detail.CreateTime, // 创建时间
			UpdateBy:   detail.UpdateBy,   // 更新者
			UpdateTime: detail.UpdateTime, // 更新时间
			ScopeType:  detail.ScopeType,  // 作用域类型
			PlatformId: detail.PlatformId, // 平台ID
			TenantId:   detail.TenantId,   // 租户ID
			MerchantId: detail.MerchantId, // 商户ID
			IsAdmin:    detail.IsAdmin,    // 是否超级管理员角色
		})
	}

	return &types.QueryRoleListResp{
		Code:     "000000",
		Message:  "查询角色成功",
		Current:  req.Current,
		Data:     list,
		PageSize: req.PageSize,
		Success:  true,
		Total:    result.Total,
	}, nil
}
