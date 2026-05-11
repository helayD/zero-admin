package digital_card_asset

import (
	"context"
	"strings"

	admincommon "github.com/feihua/zero-admin/api/admin/internal/common"
	"github.com/feihua/zero-admin/api/admin/internal/common/errorx"
	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/feihua/zero-admin/pkg/digitalcardmint"
	"github.com/zeromicro/go-zero/core/logx"
)

// AdminRevokeDigitalCardClaimTokenLogic Story 10.7 Task 8.8 / S5 — 后台手动吊销分享凭证
type AdminRevokeDigitalCardClaimTokenLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAdminRevokeDigitalCardClaimTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AdminRevokeDigitalCardClaimTokenLogic {
	return &AdminRevokeDigitalCardClaimTokenLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AdminRevokeDigitalCardClaimTokenLogic) AdminRevokeDigitalCardClaimToken(req *types.AdminRevokeDigitalCardClaimTokenReq) (*types.AdminRevokeDigitalCardClaimTokenResp, error) {
	if req.TokenId <= 0 {
		return nil, errorx.NewDefaultError("凭证ID无效")
	}
	if strings.TrimSpace(req.Reason) == "" {
		return nil, errorx.NewDefaultError("吊销原因不能为空")
	}

	writeScope, err := resolveDigitalCardAssetWriteScope(l.ctx, admincommon.RequestedGovernanceScope{
		ScopeType:  req.ScopeType,
		PlatformID: req.PlatformId,
		TenantID:   req.TenantId,
		MerchantID: req.MerchantId,
	})
	if err != nil {
		return nil, err
	}

	operatorID, err := admincommon.GetUserId(l.ctx)
	if err != nil {
		return nil, err
	}

	result, err := l.svcCtx.CardMintAdminService.AdminRevokeDigitalCardClaimToken(l.ctx, writeScope, digitalcardmint.AdminRevokeClaimTokenInput{
		TokenID:    req.TokenId,
		OperatorID: operatorID,
		Reason:     req.Reason,
		TraceID:    req.TraceId,
	})
	if err != nil {
		return nil, errorx.NewDefaultError(err.Error())
	}
	if result == nil {
		return nil, errorx.NewDefaultError("吊销失败：未返回结果")
	}

	return &types.AdminRevokeDigitalCardClaimTokenResp{
		Code:       "000000",
		Message:    result.Message,
		TokenId:    result.TokenID,
		FromStatus: result.FromStatus,
		ToStatus:   result.ToStatus,
		Success:    true,
	}, nil
}
