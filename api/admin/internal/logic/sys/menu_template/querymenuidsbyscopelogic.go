package menu_template

import (
	"context"

	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/feihua/zero-admin/rpc/sys/sysclient"
	"github.com/zeromicro/go-zero/core/logc"
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
	result, err := l.svcCtx.MenuTemplateService.QueryMenuIdsByScope(l.ctx, &sysclient.QueryMenuTemplateByScope{
		ScopeType:  trimOptionalText(req.ScopeType),
		PlatformId: req.PlatformId,
	})
	if err != nil {
		logc.Errorf(l.ctx, "按作用域查询菜单模板菜单ID失败, 参数:%+v, 异常:%s", req, err.Error())
		return nil, grpcError(err)
	}

	return &types.QueryMenuIdsByScopeResp{
		Code:    "000000",
		Message: "查询菜单模板可用菜单成功",
		Data:    result.MenuIds,
	}, nil
}
