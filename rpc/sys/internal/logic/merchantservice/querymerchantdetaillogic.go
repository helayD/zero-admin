package merchantservicelogic

import (
	"context"
	"errors"

	"github.com/feihua/zero-admin/rpc/sys/internal/svc"
	"github.com/feihua/zero-admin/rpc/sys/sysclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type QueryMerchantDetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewQueryMerchantDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryMerchantDetailLogic {
	return &QueryMerchantDetailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *QueryMerchantDetailLogic) QueryMerchantDetail(in *sysclient.QueryMerchantDetailReq) (*sysclient.QueryMerchantDetailResp, error) {
	if in == nil || in.Id <= 0 {
		return nil, errors.New("商户ID不能为空")
	}

	var row merchantQueryRow
	result := buildMerchantBaseQuery(l.svcCtx.DB.WithContext(l.ctx)).
		Select(merchantSelectColumns).
		Where("m.id = ?", in.Id).
		Limit(1).
		Scan(&row)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, errors.New("商户不存在")
	}

	return &sysclient.QueryMerchantDetailResp{
		Data: mapMerchantRowToProto(row),
	}, nil
}
