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

type EscalateChainLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewEscalateChainLogic(ctx context.Context, svcCtx *svc.ServiceContext) *EscalateChainLogic {
	return &EscalateChainLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *EscalateChainLogic) EscalateChain(req *types.EscalateChainReq) (*types.EscalateChainResp, error) {
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

	result, err := l.svcCtx.OrderService.EscalateChain(l.ctx, &omsclient.EscalateChainReq{
		OrderId:        req.OrderId,
		PlatformId:     writeScope.PlatformID,
		TenantId:       writeScope.TenantID,
		MerchantId:     writeScope.MerchantID,
		OperatorId:     operatorId,
		EscalateReason: req.EscalateReason,
	})
	if err != nil {
		logc.Errorf(l.ctx, "升级链路失败, orderId=%d, err=%s", req.OrderId, err.Error())
		s, _ := status.FromError(err)
		return nil, errorx.NewDefaultError(s.Message())
	}

	if result.Code != 0 {
		return &types.EscalateChainResp{
			Code:           strconv.FormatInt(result.Code, 10),
			Message:        result.Msg,
			ManualRequired: result.ManualRequired,
			Success:        false,
		}, nil
	}

	// 写入审计日志
	writeChainInterventionLog(l.ctx, l.svcCtx, operatorId, "escalate", req.EscalateReason, req.OrderId,
		result.BeforeStage, result.BeforeResult, result.AfterStage, result.AfterResult)

	return &types.EscalateChainResp{
		Code:           "000000",
		Message:        "链路已升级为需人工介入",
		ManualRequired: true,
		Success:        true,
	}, nil
}
