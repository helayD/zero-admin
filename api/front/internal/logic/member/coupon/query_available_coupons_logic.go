// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package coupon

import (
	"context"

	"github.com/feihua/zero-admin/api/front/internal/svc"
	"github.com/feihua/zero-admin/api/front/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type QueryAvailableCouponsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryAvailableCouponsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryAvailableCouponsLogic {
	return &QueryAvailableCouponsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QueryAvailableCouponsLogic) QueryAvailableCoupons(req *types.QueryAvailableCouponsReq) (resp *types.QueryAvailableCouponsResp, err error) {
	// todo: add your logic here and delete this line

	return
}
