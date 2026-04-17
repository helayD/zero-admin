// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package draw_activity

import (
	"context"

	"github.com/feihua/zero-admin/api/front/internal/svc"
	"github.com/feihua/zero-admin/api/front/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type QueryDrawActivityLandingLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryDrawActivityLandingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryDrawActivityLandingLogic {
	return &QueryDrawActivityLandingLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QueryDrawActivityLandingLogic) QueryDrawActivityLanding(req *types.DrawActivityLandingReq) (resp *types.DrawActivityLandingResp, err error) {
	return queryLanding(l.ctx, l.svcCtx, req, 0)
}
