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

type QueryMenuTemplateListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryMenuTemplateListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryMenuTemplateListLogic {
	return &QueryMenuTemplateListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QueryMenuTemplateListLogic) QueryMenuTemplateList(req *types.QueryMenuTemplateListReq) (resp *types.QueryMenuTemplateListResp, err error) {
	return menu_template_logic.NewQueryMenuTemplateListLogic(l.ctx, l.svcCtx).QueryMenuTemplateList(req)
}
