package menu_template

import (
	"context"

	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/feihua/zero-admin/rpc/sys/sysclient"
	"github.com/zeromicro/go-zero/core/logc"
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
	result, err := l.svcCtx.MenuTemplateService.QueryMenuTemplateList(l.ctx, &sysclient.QueryMenuTemplateListReq{
		PageNum:   req.Current,
		PageSize:  req.PageSize,
		Name:      trimOptionalText(req.Name),
		ScopeType: trimOptionalText(req.ScopeType),
		Status:    req.Status,
	})
	if err != nil {
		logc.Errorf(l.ctx, "查询菜单模板列表失败, 参数:%+v, 异常:%s", req, err.Error())
		return nil, grpcError(err)
	}

	list := make([]*types.QueryMenuTemplateListData, 0, len(result.List))
	for _, item := range result.List {
		list = append(list, mapMenuTemplateListItem(item))
	}

	return &types.QueryMenuTemplateListResp{
		Code:     "000000",
		Message:  "查询菜单模板列表成功",
		Current:  req.Current,
		Data:     list,
		PageSize: req.PageSize,
		Success:  true,
		Total:    result.Total,
	}, nil
}
