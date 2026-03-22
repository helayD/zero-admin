// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package merchant

import (
	"context"

	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/feihua/zero-admin/rpc/sys/sysclient"
	"github.com/zeromicro/go-zero/core/logc"

	"github.com/zeromicro/go-zero/core/logx"
)

type QueryMerchantDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryMerchantDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryMerchantDetailLogic {
	return &QueryMerchantDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QueryMerchantDetailLogic) QueryMerchantDetail(req *types.QueryMerchantDetailReq) (resp *types.QueryMerchantDetailResp, err error) {
	result, err := l.svcCtx.MerchantService.QueryMerchantDetail(l.ctx, &sysclient.QueryMerchantDetailReq{
		Id: req.Id,
	})
	if err != nil {
		logc.Errorf(l.ctx, "查询商户详情失败, 参数: %+v, 异常: %s", req, err.Error())
		return nil, grpcError(err)
	}

	return &types.QueryMerchantDetailResp{
		Code:    "000000",
		Message: "查询商户详情成功",
		Data:    *mapMerchantData(result.Data),
	}, nil
}
