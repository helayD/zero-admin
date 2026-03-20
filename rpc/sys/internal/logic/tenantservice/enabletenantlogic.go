package tenantservicelogic

import (
	"context"

	"github.com/feihua/zero-admin/rpc/sys/internal/svc"
	"github.com/feihua/zero-admin/rpc/sys/internal/tenantmodel"
	"github.com/feihua/zero-admin/rpc/sys/sysclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type EnableTenantLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewEnableTenantLogic(ctx context.Context, svcCtx *svc.ServiceContext) *EnableTenantLogic {
	return &EnableTenantLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *EnableTenantLogic) EnableTenant(in *sysclient.ChangeTenantStatusReq) (*sysclient.ChangeTenantStatusResp, error) {
	return changeTenantStatus(l.ctx, l.svcCtx, in, tenantmodel.TenantStatusEnabled)
}
