package digital_card_chain

import (
	"context"

	admincommon "github.com/feihua/zero-admin/api/admin/internal/common"
	"github.com/feihua/zero-admin/api/admin/internal/common/errorx"
	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type FreezeDigitalCardChainLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewFreezeDigitalCardChainLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FreezeDigitalCardChainLogic {
	return &FreezeDigitalCardChainLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *FreezeDigitalCardChainLogic) FreezeDigitalCardChain(req *types.FreezeDigitalCardChainReq) (*types.DigitalCardChainActionResp, error) {
	writeScope, err := resolveDigitalCardChainWriteScope(l.ctx, admincommon.RequestedGovernanceScope{
		ScopeType:  req.ScopeType,
		PlatformID: req.PlatformId,
		TenantID:   req.TenantId,
		MerchantID: req.MerchantId,
	})
	if err != nil {
		return nil, err
	}
	if err = validateActionReason(req.Reason); err != nil {
		return nil, err
	}
	operatorId, err := admincommon.GetUserId(l.ctx)
	if err != nil {
		return nil, errorx.NewDefaultError("无法获取操作人身份，请重新登录")
	}
	result, err := l.svcCtx.CardMintAdminService.FreezeTask(l.ctx, writeScope, req.TaskId, operatorId, req.Reason)
	if err != nil {
		return nil, errorx.NewDefaultError(err.Error())
	}
	writeDigitalCardChainOperateLog(l.ctx, l.svcCtx, operatorId, "freeze", req.TaskId, req.Reason, result)
	return &types.DigitalCardChainActionResp{
		Code:           "000000",
		Message:        "链路已冻结",
		TaskStatus:     result.TaskStatus,
		MintStatus:     result.MintStatus,
		ChainStatus:    result.ChainStatus,
		RetryCount:     result.RetryCount,
		ManualRequired: result.ManualRequired,
		Frozen:         result.Frozen,
		ChainType:      result.ChainType,
		Success:        true,
	}, nil
}
