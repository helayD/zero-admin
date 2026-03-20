package menutemplateservicelogic

import (
	"context"

	"github.com/feihua/zero-admin/rpc/sys/internal/svc"
	"github.com/feihua/zero-admin/rpc/sys/sysclient"

	"github.com/zeromicro/go-zero/core/logx"
)

type QueryMenuTemplateDetailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewQueryMenuTemplateDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryMenuTemplateDetailLogic {
	return &QueryMenuTemplateDetailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 查询菜单模板详情
func (l *QueryMenuTemplateDetailLogic) QueryMenuTemplateDetail(in *sysclient.QueryMenuTemplateDetailReq) (*sysclient.QueryMenuTemplateDetailResp, error) {
	// todo: add your logic here and delete this line

	return &sysclient.QueryMenuTemplateDetailResp{}, nil
}
