package tenantservicelogic

import (
	"context"

	"github.com/feihua/zero-admin/rpc/sys/internal/svc"
	"github.com/feihua/zero-admin/rpc/sys/internal/tenantmodel"
	"github.com/feihua/zero-admin/rpc/sys/sysclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type ArchiveTenantLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewArchiveTenantLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ArchiveTenantLogic {
	return &ArchiveTenantLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *ArchiveTenantLogic) ArchiveTenant(in *sysclient.ChangeTenantStatusReq) (*sysclient.ChangeTenantStatusResp, error) {
	return changeTenantStatus(l.ctx, l.svcCtx, in, tenantmodel.TenantStatusArchived)
}
