package menutemplateservicelogic

import (
	"context"
	"errors"

	"github.com/feihua/zero-admin/rpc/sys/internal/svc"
	"github.com/feihua/zero-admin/rpc/sys/sysclient"
	"github.com/zeromicro/go-zero/core/logc"

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
	scopeType, err := normalizeTemplateScope(in.ScopeType)
	if err != nil {
		return nil, err
	}

	menuIDs := make([]int64, 0)
	err = l.svcCtx.DB.WithContext(l.ctx).
		Table("sys_menu_template_item item").
		Distinct("item.menu_id").
		Joins("INNER JOIN sys_menu_template template ON template.id = item.template_id").
		Where("template.scope_type = ? AND template.platform_id = ? AND template.status = 1", scopeType, normalizePlatformID(in.PlatformId)).
		Order("item.menu_id ASC").
		Pluck("item.menu_id", &menuIDs).Error
	if err != nil {
		logc.Errorf(l.ctx, "根据作用域查询菜单ID失败, 参数:%+v, 异常:%s", in, err.Error())
		return nil, errors.New("查询菜单模板可用菜单失败")
	}

	return &sysclient.QueryMenuTemplateByScopeResp{MenuIds: menuIDs}, nil
}
