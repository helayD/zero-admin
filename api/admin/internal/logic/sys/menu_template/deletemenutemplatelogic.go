package menu_template

import (
	"context"

	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/feihua/zero-admin/rpc/sys/sysclient"
	"github.com/zeromicro/go-zero/core/logc"
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
	_, err = l.svcCtx.MenuTemplateService.DeleteMenuTemplate(l.ctx, &sysclient.DeleteMenuTemplateReq{
		Ids: normalizeMenuIDs(req.Ids),
	})
	if err != nil {
		logc.Errorf(l.ctx, "删除菜单模板失败, 参数:%+v, 异常:%s", req, err.Error())
		return nil, grpcError(err)
	}

	return &types.BaseResp{
		Code:    "000000",
		Message: "删除菜单模板成功",
	}, nil
}
