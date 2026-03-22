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

type UpdateMenuTemplateLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdateMenuTemplateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateMenuTemplateLogic {
	return &UpdateMenuTemplateLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdateMenuTemplateLogic) UpdateMenuTemplate(req *types.UpdateMenuTemplateReq) (resp *types.BaseResp, err error) {
	return menu_template_logic.NewUpdateMenuTemplateLogic(l.ctx, l.svcCtx).UpdateMenuTemplate(req)
}
