package menu_template

import (
	"context"

	"github.com/feihua/zero-admin/api/admin/internal/common"
	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/feihua/zero-admin/rpc/sys/sysclient"
	"github.com/zeromicro/go-zero/core/logc"
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
	userName, err := common.GetUserName(l.ctx)
	if err != nil {
		return nil, err
	}

	_, err = l.svcCtx.MenuTemplateService.UpdateMenuTemplate(l.ctx, &sysclient.UpdateMenuTemplateReq{
		Id:         req.Id,
		Name:       trimOptionalText(req.Name),
		ScopeType:  trimOptionalText(req.ScopeType),
		PlatformId: req.PlatformId,
		Status:     req.Status,
		Remark:     trimOptionalText(req.Remark),
		UpdateBy:   userName,
		MenuIds:    normalizeMenuIDs(req.MenuIds),
	})
	if err != nil {
		logc.Errorf(l.ctx, "更新菜单模板失败, 参数:%+v, 异常:%s", req, err.Error())
		return nil, grpcError(err)
	}

	return &types.BaseResp{
		Code:    "000000",
		Message: "更新菜单模板成功",
	}, nil
}
