package merchantservicelogic

import (
	"context"
	"errors"

	"github.com/feihua/zero-admin/rpc/sys/internal/svc"
	"github.com/feihua/zero-admin/rpc/sys/sysclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type QueryMerchantListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewQueryMerchantListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryMerchantListLogic {
	return &QueryMerchantListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *QueryMerchantListLogic) QueryMerchantList(in *sysclient.QueryMerchantListReq) (*sysclient.QueryMerchantListResp, error) {
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

	baseQuery := applyMerchantFilters(buildMerchantBaseQuery(l.svcCtx.DB.WithContext(l.ctx)), in)
	var total int64
	if err := baseQuery.Distinct("m.id").Count(&total).Error; err != nil {
		return nil, err
	}

	var rows []merchantQueryRow
	if err := applyMerchantFilters(buildMerchantBaseQuery(l.svcCtx.DB.WithContext(l.ctx)), in).
		Select(merchantSelectColumns).
		Order("m.id DESC").
		Limit(int(pageSize)).
		Offset(int((pageNum - 1) * pageSize)).
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	list := make([]*sysclient.MerchantData, 0, len(rows))
	for _, row := range rows {
		list = append(list, mapMerchantRowToProto(row))
	}

	return &sysclient.QueryMerchantListResp{
		Total: total,
		List:  list,
	}, nil
}
