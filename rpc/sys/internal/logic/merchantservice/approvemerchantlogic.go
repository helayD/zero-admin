package merchantservicelogic

import (
	"context"

	"github.com/feihua/zero-admin/rpc/sys/internal/merchantmodel"
	"github.com/feihua/zero-admin/rpc/sys/internal/svc"
	"github.com/feihua/zero-admin/rpc/sys/sysclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type ApproveMerchantLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewApproveMerchantLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ApproveMerchantLogic {
	return &ApproveMerchantLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ApproveMerchantLogic) ApproveMerchant(in *sysclient.ReviewMerchantReq) (*sysclient.ReviewMerchantResp, error) {
	return changeMerchantReviewStatus(l.ctx, l.svcCtx, in, merchantmodel.MerchantReviewApproved)
}
