package merchantservicelogic

import (
	"context"

	"github.com/feihua/zero-admin/rpc/sys/internal/merchantmodel"
	"github.com/feihua/zero-admin/rpc/sys/internal/svc"
	"github.com/feihua/zero-admin/rpc/sys/sysclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type DisableMerchantLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDisableMerchantLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DisableMerchantLogic {
	return &DisableMerchantLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DisableMerchantLogic) DisableMerchant(in *sysclient.ChangeMerchantStatusReq) (*sysclient.ChangeMerchantStatusResp, error) {
	return changeMerchantBusinessStatus(l.ctx, l.svcCtx, in, merchantmodel.MerchantBusinessDisabled)
}
