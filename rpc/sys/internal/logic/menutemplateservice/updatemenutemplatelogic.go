package menutemplateservicelogic

import (
	"context"

	"github.com/feihua/zero-admin/rpc/sys/internal/svc"
	"github.com/feihua/zero-admin/rpc/sys/sysclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdateMenuTemplateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateMenuTemplateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateMenuTemplateLogic {
	return &UpdateMenuTemplateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 更新菜单模板
func (l *UpdateMenuTemplateLogic) UpdateMenuTemplate(in *sysclient.UpdateMenuTemplateReq) (*sysclient.UpdateMenuTemplateResp, error) {
	// todo: add your logic here and delete this line

	return &sysclient.UpdateMenuTemplateResp{}, nil
}
