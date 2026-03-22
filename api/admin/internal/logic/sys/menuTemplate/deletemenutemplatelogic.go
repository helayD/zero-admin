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

type DeleteMenuTemplateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteMenuTemplateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteMenuTemplateLogic {
	return &DeleteMenuTemplateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteMenuTemplateLogic) DeleteMenuTemplate(req *types.DeleteMenuTemplateReq) (resp *types.BaseResp, err error) {
	return menu_template_logic.NewDeleteMenuTemplateLogic(l.ctx, l.svcCtx).DeleteMenuTemplate(req)
}
