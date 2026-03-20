package menutemplateservicelogic

import (
	"context"

	"github.com/feihua/zero-admin/rpc/sys/internal/svc"
	"github.com/feihua/zero-admin/rpc/sys/sysclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type AddMenuTemplateLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewAddMenuTemplateLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddMenuTemplateLogic {
	return &AddMenuTemplateLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 添加菜单模板
func (l *AddMenuTemplateLogic) AddMenuTemplate(in *sysclient.AddMenuTemplateReq) (*sysclient.AddMenuTemplateResp, error) {
	// todo: add your logic here and delete this line

	return &sysclient.AddMenuTemplateResp{}, nil
}
