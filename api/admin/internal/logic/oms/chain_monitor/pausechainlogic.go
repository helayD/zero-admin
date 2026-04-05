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

type PauseChainLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPauseChainLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PauseChainLogic {
	return &PauseChainLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PauseChainLogic) PauseChain(req *types.PauseChainReq) (*types.PauseChainResp, error) {
	writeScope, err := resolveChainWriteScope(l.ctx, admincommon.RequestedGovernanceScope{
		ScopeType:  req.ScopeType,
		PlatformID: req.PlatformId,
		TenantID:   req.TenantId,
		MerchantID: req.MerchantId,
	})
	if err != nil {
		return nil, err
	}

	operatorId, err := admincommon.GetUserId(l.ctx)
	if err != nil {
		logc.Errorf(l.ctx, "获取操作人ID失败: %s", err.Error())
		return nil, errorx.NewDefaultError("无法获取操作人身份，请重新登录")
	}

	result, err := l.svcCtx.OrderService.PauseCompensationChain(l.ctx, &omsclient.PauseCompensationChainReq{
		OrderId:     req.OrderId,
		PlatformId:  writeScope.PlatformID,
		TenantId:    writeScope.TenantID,
		MerchantId:  writeScope.MerchantID,
		OperatorId:  operatorId,
		PauseReason: req.PauseReason,
	})
	if err != nil {
		logc.Errorf(l.ctx, "暂停链路失败, orderId=%d, err=%s", req.OrderId, err.Error())
		s, _ := status.FromError(err)
		return nil, errorx.NewDefaultError(s.Message())
	}

	if result.Code != 0 {
		return &types.PauseChainResp{
			Code:    strconv.FormatInt(result.Code, 10),
			Message: result.Msg,
			Paused:  result.Paused,
			Success: false,
		}, nil
	}

	// 写入审计日志
	writeChainInterventionLog(l.ctx, l.svcCtx, operatorId, "pause", req.PauseReason, req.OrderId,
		result.BeforeStage, result.BeforeResult, result.AfterStage, result.AfterResult)

	return &types.PauseChainResp{
		Code:    "000000",
		Message: "链路已暂停",
		Paused:  true,
		Success: true,
	}, nil
}
