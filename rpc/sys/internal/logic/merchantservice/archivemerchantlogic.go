package merchantservicelogic

import (
	"context"

	"github.com/feihua/zero-admin/rpc/sys/internal/merchantmodel"
	"github.com/feihua/zero-admin/rpc/sys/internal/svc"
	"github.com/feihua/zero-admin/rpc/sys/sysclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type ArchiveMerchantLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewArchiveMerchantLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ArchiveMerchantLogic {
	return &ArchiveMerchantLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ArchiveMerchantLogic) ArchiveMerchant(in *sysclient.ChangeMerchantStatusReq) (*sysclient.ChangeMerchantStatusResp, error) {
	return changeMerchantBusinessStatus(l.ctx, l.svcCtx, in, merchantmodel.MerchantBusinessArchived)
}
