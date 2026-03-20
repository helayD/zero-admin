// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package tenant

import (
	"context"
	"strings"

	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/feihua/zero-admin/rpc/sys/sysclient"
	"github.com/zeromicro/go-zero/core/logc"

	"github.com/zeromicro/go-zero/core/logx"
)

type QueryTenantListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryTenantListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryTenantListLogic {
	return &QueryTenantListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QueryTenantListLogic) QueryTenantList(req *types.QueryTenantListReq) (resp *types.QueryTenantListResp, err error) {
	result, err := l.svcCtx.TenantService.QueryTenantList(l.ctx, &sysclient.QueryTenantListReq{
		TenantName: strings.TrimSpace(req.TenantName),
		TenantCode: strings.TrimSpace(req.TenantCode),
		Status:     req.Status,
		Channel:    strings.TrimSpace(req.Channel),
		PageNum:    req.Current,
		PageSize:   req.PageSize,
	})
	if err != nil {
		logc.Errorf(l.ctx, "查询租户列表失败, 参数: %+v, 异常: %s", req, err.Error())
		return nil, grpcError(err)
	}

	list := make([]*types.TenantData, 0, len(result.List))
	for _, item := range result.List {
		list = append(list, mapTenantData(item))
	}

	return &types.QueryTenantListResp{
		Code:     "000000",
		Message:  "查询租户列表成功",
		Current:  req.Current,
		Data:     list,
		PageSize: req.PageSize,
		Success:  true,
		Total:    result.Total,
	}, nil
}
