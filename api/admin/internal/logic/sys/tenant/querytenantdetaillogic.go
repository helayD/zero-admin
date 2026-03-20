// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package tenant

import (
	"context"

	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/feihua/zero-admin/rpc/sys/sysclient"
	"github.com/zeromicro/go-zero/core/logc"

	"github.com/zeromicro/go-zero/core/logx"
)

type QueryTenantDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryTenantDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryTenantDetailLogic {
	return &QueryTenantDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QueryTenantDetailLogic) QueryTenantDetail(req *types.QueryTenantDetailReq) (resp *types.QueryTenantDetailResp, err error) {
	result, err := l.svcCtx.TenantService.QueryTenantDetail(l.ctx, &sysclient.QueryTenantDetailReq{
		Id: req.Id,
	})
	if err != nil {
		logc.Errorf(l.ctx, "查询租户详情失败, 参数: %+v, 异常: %s", req, err.Error())
		return nil, grpcError(err)
	}

	return &types.QueryTenantDetailResp{
		Code:    "000000",
		Message: "查询租户详情成功",
		Data:    *mapTenantData(result.Data),
	}, nil
}
