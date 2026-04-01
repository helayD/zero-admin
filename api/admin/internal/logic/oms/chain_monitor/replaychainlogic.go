package chain_monitor

import (
	"context"
	"strconv"

	admincommon "github.com/feihua/zero-admin/api/admin/internal/common"
	"github.com/feihua/zero-admin/api/admin/internal/common/errorx"
	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/feihua/zero-admin/rpc/oms/omsclient"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/status"
)

type ReplayChainLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewReplayChainLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ReplayChainLogic {
	return &ReplayChainLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ReplayChainLogic) ReplayChain(req *types.ReplayChainReq) (*types.ReplayChainResp, error) {
	current, err := admincommon.CurrentGovernanceScope(l.ctx)
	if err != nil {
		return nil, errorx.NewDefaultError(err.Error())
	}

	operatorId, err := admincommon.GetUserId(l.ctx)
	if err != nil {
		logc.Errorf(l.ctx, "获取操作人ID失败: %s", err.Error())
		return nil, errorx.NewDefaultError("无法获取操作人身份，请重新登录")
	}

	result, err := l.svcCtx.OrderService.ReplayCompensationChain(l.ctx, &omsclient.ReplayCompensationChainReq{
		OrderId:      req.OrderId,
		PlatformId:   current.PlatformID,
		TenantId:     current.TenantID,
		MerchantId:   current.MerchantID,
		OperatorId:   operatorId,
		ReplayReason: req.ReplayReason,
	})
	if err != nil {
		logc.Errorf(l.ctx, "回放链路失败, orderId=%d, err=%s", req.OrderId, err.Error())
		s, _ := status.FromError(err)
		return nil, errorx.NewDefaultError(s.Message())
	}

	if result.Code != 0 {
		return &types.ReplayChainResp{
			Code:    strconv.FormatInt(result.Code, 10),
			Message: result.Msg,
			TraceId: result.TraceId,
			Success: false,
		}, nil
	}

	// 写入审计日志
	writeChainInterventionLog(l.ctx, l.svcCtx, operatorId, "replay", req.ReplayReason, req.OrderId,
		result.BeforeStage, result.BeforeResult, result.AfterStage, result.AfterResult)

	return &types.ReplayChainResp{
		Code:    "000000",
		Message: "链路回放成功",
		TraceId: result.TraceId,
		Success: true,
	}, nil
}
