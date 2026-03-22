package menu_template

import (
	"context"

	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/feihua/zero-admin/rpc/sys/sysclient"
	"github.com/zeromicro/go-zero/core/logc"
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
	result, err := l.svcCtx.MenuTemplateService.QueryMenuTemplateDetail(l.ctx, &sysclient.QueryMenuTemplateDetailReq{
		Id: req.Id,
	})
	if err != nil {
		logc.Errorf(l.ctx, "查询菜单模板详情失败, 参数:%+v, 异常:%s", req, err.Error())
		return nil, grpcError(err)
	}

	return &types.QueryMenuTemplateDetailResp{
		Code:    "000000",
		Message: "查询菜单模板详情成功",
		Data:    mapMenuTemplateDetail(result),
	}, nil
}
