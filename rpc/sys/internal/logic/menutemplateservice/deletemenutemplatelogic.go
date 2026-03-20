package menutemplateservicelogic

import (
	"context"

	"github.com/feihua/zero-admin/rpc/sys/internal/svc"
	"github.com/feihua/zero-admin/rpc/sys/sysclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteMenuTemplateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteMenuTemplateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteMenuTemplateLogic {
	return &DeleteMenuTemplateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 删除菜单模板
func (l *DeleteMenuTemplateLogic) DeleteMenuTemplate(in *sysclient.DeleteMenuTemplateReq) (*sysclient.DeleteMenuTemplateResp, error) {
	// todo: add your logic here and delete this line

	return &sysclient.DeleteMenuTemplateResp{}, nil
}
