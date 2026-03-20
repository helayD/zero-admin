package tenantservicelogic

import (
	"context"
	"errors"

	"github.com/feihua/zero-admin/rpc/sys/internal/svc"
	"github.com/feihua/zero-admin/rpc/sys/sysclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type QueryTenantDetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewQueryTenantDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryTenantDetailLogic {
	return &QueryTenantDetailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *QueryTenantDetailLogic) QueryTenantDetail(in *sysclient.QueryTenantDetailReq) (*sysclient.QueryTenantDetailResp, error) {
	if in == nil || in.Id <= 0 {
		return nil, errors.New("租户ID不能为空")
	}

	var row tenantQueryRow
	result := buildTenantBaseQuery(l.svcCtx.DB.WithContext(l.ctx)).
		Select(tenantSelectColumns).
		Where("t.id = ?", in.Id).
		Limit(1).
		Scan(&row)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, errors.New("租户不存在")
	}

	return &sysclient.QueryTenantDetailResp{
		Data: mapTenantRowToProto(row),
	}, nil
}
