package digital_card_asset

import (
	"context"

	admincommon "github.com/feihua/zero-admin/api/admin/internal/common"
	"github.com/feihua/zero-admin/api/admin/internal/common/errorx"
	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/zeromicro/go-zero/core/logx"
)

type OfflineDigitalCardAssetDisplayLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewOfflineDigitalCardAssetDisplayLogic(ctx context.Context, svcCtx *svc.ServiceContext) *OfflineDigitalCardAssetDisplayLogic {
	return &OfflineDigitalCardAssetDisplayLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *OfflineDigitalCardAssetDisplayLogic) OfflineDigitalCardAssetDisplay(req *types.DigitalCardAssetActionReq) (*types.DigitalCardAssetActionResp, error) {
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

	result, err := l.svcCtx.CardMintAdminService.OfflineAssetDisplay(l.ctx, writeScope, req.AssetInstanceId, operatorID, req.Reason)
	if err != nil {
		return nil, errorx.NewDefaultError(err.Error())
	}

	return &types.DigitalCardAssetActionResp{
		Code:                 "000000",
		Message:              "数字卡片资产已下线展示",
		AssetInstanceId:      result.AssetInstanceID,
		DisplayStatus:        result.DisplayStatus,
		DisplayStatusText:    result.DisplayStatusText,
		ComplianceStatus:     result.ComplianceStatus,
		ComplianceStatusText: result.ComplianceStatusText,
		Success:              true,
	}, nil
}
