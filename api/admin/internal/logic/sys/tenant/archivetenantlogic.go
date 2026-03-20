// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package tenant

import (
	"context"
	"strings"

	"github.com/feihua/zero-admin/api/admin/internal/common"
	"github.com/feihua/zero-admin/api/admin/internal/common/res"
	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/feihua/zero-admin/rpc/sys/sysclient"
	"github.com/zeromicro/go-zero/core/logc"

	"github.com/zeromicro/go-zero/core/logx"
)

type ArchiveTenantLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewArchiveTenantLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ArchiveTenantLogic {
	return &ArchiveTenantLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ArchiveTenantLogic) ArchiveTenant(req *types.ChangeTenantStatusReq) (resp *types.BaseResp, err error) {
	userName, err := common.GetUserName(l.ctx)
	if err != nil {
		return nil, err
	}
	userID, err := common.GetUserId(l.ctx)
	if err != nil {
		return nil, err
	}

	_, err = l.svcCtx.TenantService.ArchiveTenant(l.ctx, &sysclient.ChangeTenantStatusReq{
		Ids:          req.Ids,
		StatusReason: strings.TrimSpace(req.StatusReason),
		UpdateBy:     userName,
		OperatorId:   userID,
	})
	if err != nil {
		logc.Errorf(l.ctx, "归档租户失败, 参数: %+v, 异常: %s", req, err.Error())
		return nil, grpcError(err)
	}

	return res.Success()
}
