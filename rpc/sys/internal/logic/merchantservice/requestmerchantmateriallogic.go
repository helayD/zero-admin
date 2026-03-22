package merchantservicelogic

import (
	"context"

	"github.com/feihua/zero-admin/rpc/sys/internal/merchantmodel"
	"github.com/feihua/zero-admin/rpc/sys/internal/svc"
	"github.com/feihua/zero-admin/rpc/sys/sysclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type RequestMerchantMaterialLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRequestMerchantMaterialLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RequestMerchantMaterialLogic {
	return &RequestMerchantMaterialLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RequestMerchantMaterialLogic) RequestMerchantMaterial(in *sysclient.ReviewMerchantReq) (*sysclient.ReviewMerchantResp, error) {
	return changeMerchantReviewStatus(l.ctx, l.svcCtx, in, merchantmodel.MerchantReviewMaterialRequired)
}
