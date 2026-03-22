// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package menuTemplate

import (
	"context"

	menu_template_logic "github.com/feihua/zero-admin/api/admin/internal/logic/sys/menu_template"
	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type QueryMenuIdsByScopeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryMenuIdsByScopeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryMenuIdsByScopeLogic {
	return &QueryMenuIdsByScopeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QueryMenuIdsByScopeLogic) QueryMenuIdsByScope(req *types.QueryMenuIdsByScopeReq) (resp *types.QueryMenuIdsByScopeResp, err error) {
	return menu_template_logic.NewQueryMenuIdsByScopeLogic(l.ctx, l.svcCtx).QueryMenuIdsByScope(req)
}
