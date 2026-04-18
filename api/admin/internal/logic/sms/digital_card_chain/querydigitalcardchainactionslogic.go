package digital_card_chain

import (
	"context"

	admincommon "github.com/feihua/zero-admin/api/admin/internal/common"
	"github.com/feihua/zero-admin/api/admin/internal/common/errorx"
	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type QueryDigitalCardChainActionsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryDigitalCardChainActionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryDigitalCardChainActionsLogic {
	return &QueryDigitalCardChainActionsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QueryDigitalCardChainActionsLogic) QueryDigitalCardChainActions(req *types.QueryDigitalCardChainActionsReq) (*types.QueryDigitalCardChainActionsResp, error) {
	current, err := admincommon.CurrentGovernanceScope(l.ctx)
	if err != nil {
		return nil, errorx.NewDefaultError(err.Error())
	}
	actions, err := l.svcCtx.CardMintService.QueryAvailableActions(l.ctx, current, req.TaskId)
	if err != nil {
		return nil, errorx.NewDefaultError(err.Error())
	}
	return &types.QueryDigitalCardChainActionsResp{
		Code:             "000000",
		Message:          "查询成功",
		AvailableActions: actions,
		Success:          true,
	}, nil
}
