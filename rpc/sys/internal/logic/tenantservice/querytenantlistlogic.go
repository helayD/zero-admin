package tenantservicelogic

import (
	"context"
	"errors"

	"github.com/feihua/zero-admin/rpc/sys/internal/svc"
	"github.com/feihua/zero-admin/rpc/sys/sysclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type QueryTenantListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewQueryTenantListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryTenantListLogic {
	return &QueryTenantListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *QueryTenantListLogic) QueryTenantList(in *sysclient.QueryTenantListReq) (*sysclient.QueryTenantListResp, error) {
	if in == nil {
		return nil, errors.New("请求不能为空")
	}
	pageNum := in.PageNum
	if pageNum <= 0 {
		pageNum = 1
	}
	pageSize := in.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	baseQuery := applyTenantFilters(buildTenantBaseQuery(l.svcCtx.DB.WithContext(l.ctx)), in)
	var total int64
	if err := baseQuery.Distinct("t.id").Count(&total).Error; err != nil {
		return nil, err
	}

	var rows []tenantQueryRow
	if err := applyTenantFilters(buildTenantBaseQuery(l.svcCtx.DB.WithContext(l.ctx)), in).
		Select(tenantSelectColumns).
		Order("t.id DESC").
		Limit(int(pageSize)).
		Offset(int((pageNum - 1) * pageSize)).
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	list := make([]*sysclient.TenantData, 0, len(rows))
	for _, row := range rows {
		list = append(list, mapTenantRowToProto(row))
	}

	return &sysclient.QueryTenantListResp{
		Total: total,
		List:  list,
	}, nil
}
