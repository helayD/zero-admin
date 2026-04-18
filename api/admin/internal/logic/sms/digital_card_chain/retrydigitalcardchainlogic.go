package digital_card_chain

import (
	"context"

	admincommon "github.com/feihua/zero-admin/api/admin/internal/common"
	"github.com/feihua/zero-admin/api/admin/internal/common/errorx"
	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type RetryDigitalCardChainLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRetryDigitalCardChainLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RetryDigitalCardChainLogic {
	return &RetryDigitalCardChainLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RetryDigitalCardChainLogic) RetryDigitalCardChain(req *types.RetryDigitalCardChainReq) (*types.DigitalCardChainActionResp, error) {
	writeScope, err := resolveDigitalCardChainWriteScope(l.ctx, admincommon.RequestedGovernanceScope{
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
		return nil, errorx.NewDefaultError("无法获取操作人身份，请重新登录")
	}
	result, err := l.svcCtx.CardMintService.RetryTask(l.ctx, writeScope, req.TaskId, operatorId, req.Reason)
	if err != nil {
		return nil, errorx.NewDefaultError(err.Error())
	}
	writeDigitalCardChainOperateLog(l.ctx, l.svcCtx, operatorId, "retry", req.TaskId, req.Reason, result)
	return &types.DigitalCardChainActionResp{
		Code:           "000000",
		Message:        "链路重试已触发",
		TaskStatus:     result.TaskStatus,
		MintStatus:     result.MintStatus,
		ChainStatus:    result.ChainStatus,
		RetryCount:     result.RetryCount,
		ManualRequired: result.ManualRequired,
		Frozen:         result.Frozen,
		Success:        true,
	}, nil
}
