package dept

import (
	"context"
	admincommon "github.com/feihua/zero-admin/api/admin/internal/common"
	"github.com/feihua/zero-admin/api/admin/internal/common/errorx"
	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/feihua/zero-admin/rpc/sys/sysclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/status"
)

// QueryDeptListLogic 部门列表
/*
Author: LiuFeiHua
Date: 2023/12/18 17:16
*/
type QueryDeptListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryDeptListLogic(ctx context.Context, svcCtx *svc.ServiceContext) QueryDeptListLogic {
	return QueryDeptListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// QueryDeptList 查询部门列表
func (l *QueryDeptListLogic) QueryDeptList(req *types.QueryDeptListReq) (*types.QueryDeptListResp, error) {
	scopeReq, err := admincommon.BuildGovernanceScope(req.ScopeType, req.PlatformId, req.TenantId, req.MerchantId)
	if err != nil {
		return nil, errorx.NewDefaultError(err.Error())
	}

	resp, err := l.svcCtx.DeptService.QueryDeptList(l.ctx, &sysclient.QueryDeptListReq{Scope: scopeReq})

	if err != nil {
		logc.Errorf(l.ctx, "查询部门列表,参数: %+v,异常:%s", req, err.Error())
		s, _ := status.FromError(err)
		return nil, errorx.NewDefaultError(s.Message())
	}

	var list []*types.QueryDeptListData

	for _, dept := range resp.List {
		item := &types.QueryDeptListData{
			Id:         dept.Id,         // 部门id
			ParentId:   dept.ParentId,   // 上级部门id
			Ancestors:  dept.Ancestors,  // 祖级列表
			DeptName:   dept.DeptName,   // 部门名称
			Sort:       dept.Sort,       // 显示顺序
			Leader:     dept.Leader,     // 负责人
			Phone:      dept.Phone,      // 联系电话
			Email:      dept.Email,      // 邮箱
			Status:     dept.Status,     // 部门状态（0：停用，1:正常）
			DelFlag:    dept.DelFlag,    // 删除标志（0代表删除 1代表存在）
			Remark:     dept.Remark,     // 备注信息
			CreateBy:   dept.CreateBy,   // 创建者
			CreateTime: dept.CreateTime, // 创建时间
			UpdateBy:   dept.UpdateBy,   // 更新者
			UpdateTime: dept.UpdateTime, // 更新时间
		}
		item.ScopeType, item.ScopeLabel, item.PlatformId, item.TenantId, item.MerchantId = admincommon.ReadGovernanceScope(dept.Scope)
		list = append(list, item)
	}

	return &types.QueryDeptListResp{
		Code:    "000000",
		Message: "查询部门成功",
		Data:    list,
		Success: true,
	}, nil
}
