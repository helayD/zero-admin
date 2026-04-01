package chain_monitor

import (
	"context"

	admincommon "github.com/feihua/zero-admin/api/admin/internal/common"
	"github.com/feihua/zero-admin/api/admin/internal/common/errorx"
	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/feihua/zero-admin/rpc/oms/omsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"google.golang.org/grpc/status"

	"github.com/zeromicro/go-zero/core/logx"
)

type QueryChainActionsLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryChainActionsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryChainActionsLogic {
	return &QueryChainActionsLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QueryChainActionsLogic) QueryChainActions(req *types.ChainActionsReq) (*types.ChainActionsResp, error) {
	current, err := admincommon.CurrentGovernanceScope(l.ctx)
	if err != nil {
		return nil, errorx.NewDefaultError(err.Error())
	}

	result, err := l.svcCtx.OrderService.QueryChainActions(l.ctx, &omsclient.QueryChainActionsReq{
		OrderId:   req.OrderId,
		PlatformId: current.PlatformID,
		TenantId:  current.TenantID,
		MerchantId: current.MerchantID,
	})
	if err != nil {
		logc.Errorf(l.ctx, "查询可用干预动作失败, orderId=%d, err=%s", req.OrderId, err.Error())
		s, _ := status.FromError(err)
		return nil, errorx.NewDefaultError(s.Message())
	}

	if result.Code != 0 {
		return &types.ChainActionsResp{
			Code:             "0",
			Message:          result.Msg,
			AvailableActions: []string{},
			Success:         false,
		}, nil
	}

	return &types.ChainActionsResp{
		Code:             "000000",
		Message:          "查询成功",
		AvailableActions: result.AvailableActions,
		Success:          true,
	}, nil
}
