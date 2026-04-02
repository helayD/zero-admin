// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package order

import (
	"context"

	"github.com/feihua/zero-admin/api/front/internal/svc"
	"github.com/feihua/zero-admin/api/front/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type QueryLogisticsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryLogisticsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryLogisticsLogic {
	return &QueryLogisticsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QueryLogisticsLogic) QueryLogistics(req *types.QueryLogisticsReq) (resp *types.QueryLogisticsResp, err error) {
	// todo: add your logic here and delete this line

	return
}
