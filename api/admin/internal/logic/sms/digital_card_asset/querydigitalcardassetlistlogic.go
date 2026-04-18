package digital_card_asset

import (
	"context"

	admincommon "github.com/feihua/zero-admin/api/admin/internal/common"
	"github.com/feihua/zero-admin/api/admin/internal/common/errorx"
	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/feihua/zero-admin/pkg/digitalcardmint"
	"github.com/zeromicro/go-zero/core/logx"
)

type QueryDigitalCardAssetListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryDigitalCardAssetListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryDigitalCardAssetListLogic {
	return &QueryDigitalCardAssetListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QueryDigitalCardAssetListLogic) QueryDigitalCardAssetList(req *types.QueryDigitalCardAssetListReq) (*types.QueryDigitalCardAssetListResp, error) {
	queryScope, err := admincommon.ResolveQueryGovernanceScope(l.ctx, admincommon.RequestedGovernanceScope{
		ScopeType:  req.ScopeType,
		PlatformID: req.PlatformId,
		TenantID:   req.TenantId,
		MerchantID: req.MerchantId,
	})
	if err != nil {
		return nil, errorx.NewDefaultError(err.Error())
	}

	total, list, err := l.svcCtx.CardMintAdminService.QueryAssetAuditList(l.ctx, queryScope, digitalcardmint.DigitalCardAssetAuditFilter{
		PageNum:          int32(req.Current),
		PageSize:         int32(req.PageSize),
		ActivityID:       req.ActivityId,
		ActivityName:     req.ActivityName,
		MemberID:         req.MemberId,
		TemplateID:       req.TemplateId,
		TemplateName:     req.TemplateName,
		AssetNo:          req.AssetNo,
		TokenID:          req.TokenId,
		MintStatus:       req.MintStatus,
		ChainStatus:      req.ChainStatus,
		DisplayStatus:    req.DisplayStatus,
		ComplianceStatus: req.ComplianceStatus,
		StartTime:        req.StartTime,
		EndTime:          req.EndTime,
	})
	if err != nil {
		return nil, errorx.NewDefaultError(err.Error())
	}

	items := make([]*types.DigitalCardAssetItem, 0, len(list))
	for index := range list {
		items = append(items, mapAuditItem(&list[index]))
	}

	return &types.QueryDigitalCardAssetListResp{
		Code:    "000000",
		Message: "查询数字卡片资产成功",
		Data: types.QueryDigitalCardAssetListData{
			List:  items,
			Total: total,
		},
		Current:  req.Current,
		PageSize: req.PageSize,
		Total:    total,
		Success:  true,
	}, nil
}
