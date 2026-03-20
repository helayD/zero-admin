package menutemplateservicelogic

import (
	"context"

	"github.com/feihua/zero-admin/rpc/sys/internal/svc"
	"github.com/feihua/zero-admin/rpc/sys/sysclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type QueryMenuTemplateListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewQueryMenuTemplateListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryMenuTemplateListLogic {
	return &QueryMenuTemplateListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查询菜单模板列表
func (l *QueryMenuTemplateListLogic) QueryMenuTemplateList(in *sysclient.QueryMenuTemplateListReq) (*sysclient.QueryMenuTemplateListResp, error) {
	// todo: add your logic here and delete this line

	return &sysclient.QueryMenuTemplateListResp{}, nil
}
