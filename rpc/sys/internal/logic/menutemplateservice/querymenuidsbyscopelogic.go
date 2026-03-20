package menutemplateservicelogic

import (
	"context"

	"github.com/feihua/zero-admin/rpc/sys/internal/svc"
	"github.com/feihua/zero-admin/rpc/sys/sysclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type QueryMenuIdsByScopeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewQueryMenuIdsByScopeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryMenuIdsByScopeLogic {
	return &QueryMenuIdsByScopeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 根据作用域查询可用菜单ID
func (l *QueryMenuIdsByScopeLogic) QueryMenuIdsByScope(in *sysclient.QueryMenuTemplateByScope) (*sysclient.QueryMenuTemplateByScopeResp, error) {
	// todo: add your logic here and delete this line

	return &sysclient.QueryMenuTemplateByScopeResp{}, nil
}
