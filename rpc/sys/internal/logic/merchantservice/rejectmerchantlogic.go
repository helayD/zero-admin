package merchantservicelogic

import (
	"context"

	"github.com/feihua/zero-admin/rpc/sys/internal/merchantmodel"
	"github.com/feihua/zero-admin/rpc/sys/internal/svc"
	"github.com/feihua/zero-admin/rpc/sys/sysclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type RejectMerchantLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRejectMerchantLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RejectMerchantLogic {
	return &RejectMerchantLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RejectMerchantLogic) RejectMerchant(in *sysclient.ReviewMerchantReq) (*sysclient.ReviewMerchantResp, error) {
	return changeMerchantReviewStatus(l.ctx, l.svcCtx, in, merchantmodel.MerchantReviewRejected)
}
