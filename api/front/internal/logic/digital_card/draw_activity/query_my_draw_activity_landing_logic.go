// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package draw_activity

import (
	"context"

	"github.com/feihua/zero-admin/api/front/internal/svc"
	"github.com/feihua/zero-admin/api/front/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type QueryMyDrawActivityLandingLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryMyDrawActivityLandingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryMyDrawActivityLandingLogic {
	return &QueryMyDrawActivityLandingLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QueryMyDrawActivityLandingLogic) QueryMyDrawActivityLanding(req *types.DrawActivityLandingReq) (resp *types.DrawActivityLandingResp, err error) {
	memberID, err := currentMemberID(l.ctx)
	if err != nil {
		return nil, err
	}
	return queryLanding(l.ctx, l.svcCtx, req, memberID)
}
