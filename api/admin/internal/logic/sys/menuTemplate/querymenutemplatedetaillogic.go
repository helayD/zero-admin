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

type QueryMenuTemplateDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryMenuTemplateDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryMenuTemplateDetailLogic {
	return &QueryMenuTemplateDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QueryMenuTemplateDetailLogic) QueryMenuTemplateDetail(req *types.QueryMenuTemplateDetailReq) (resp *types.QueryMenuTemplateDetailResp, err error) {
	return menu_template_logic.NewQueryMenuTemplateDetailLogic(l.ctx, l.svcCtx).QueryMenuTemplateDetail(req)
}
