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

// QueryDigitalCardClaimTokenListLogic Story 10.7 Task 8.8 / S5 — 后台分享凭证管理
type QueryDigitalCardClaimTokenListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryDigitalCardClaimTokenListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryDigitalCardClaimTokenListLogic {
	return &QueryDigitalCardClaimTokenListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QueryDigitalCardClaimTokenListLogic) QueryDigitalCardClaimTokenList(req *types.QueryDigitalCardClaimTokenListReq) (*types.QueryDigitalCardClaimTokenListResp, error) {
	queryScope, err := admincommon.ResolveQueryGovernanceScope(l.ctx, admincommon.RequestedGovernanceScope{
		ScopeType:  req.ScopeType,
		PlatformID: req.PlatformId,
		TenantID:   req.TenantId,
		MerchantID: req.MerchantId,
	})
	if err != nil {
		return nil, errorx.NewDefaultError(err.Error())
	}

	pageNum := int64(req.Current)
	if pageNum <= 0 {
		pageNum = 1
	}
	pageSize := int64(req.PageSize)
	if pageSize <= 0 {
		pageSize = 20
	}

	total, list, err := l.svcCtx.CardMintAdminService.QueryDigitalCardClaimTokenList(l.ctx, queryScope, digitalcardmint.DigitalCardClaimTokenFilter{
		PageNum:        pageNum,
		PageSize:       pageSize,
		TokenID:        req.TokenId,
		CardInstanceID: req.CardInstanceId,
		AssetNo:        req.AssetNo,
		IssuerID:       req.IssuerId,
		Status:         req.Status,
		DateFrom:       req.DateFrom,
		DateTo:         req.DateTo,
	})
	if err != nil {
		return nil, errorx.NewDefaultError(err.Error())
	}

	items := make([]types.DigitalCardClaimTokenListItem, 0, len(list))
	for _, item := range list {
		items = append(items, types.DigitalCardClaimTokenListItem{
			Id:             item.ID,
			TokenMasked:    item.TokenMasked,
			CardInstanceId: item.CardInstanceID,
			AssetNo:        item.AssetNo,
			TemplateName:   item.TemplateName,
			IssuerId:       item.IssuerID,
			IssuerType:     item.IssuerType,
			ExpireAt:       item.ExpireAt,
			MaxClaims:      item.MaxClaims,
			ClaimedCount:   item.ClaimedCount,
			Status:         item.Status,
			ClaimedBy:      item.ClaimedBy,
			ClaimedAt:      item.ClaimedAt,
			PlatformId:     item.PlatformID,
			TenantId:       item.TenantID,
			MerchantId:     item.MerchantID,
			CreateTime:     item.CreatedAt,
			UpdateTime:     item.UpdatedAt,
		})
	}

	return &types.QueryDigitalCardClaimTokenListResp{
		Code:     "000000",
		Message:  "查询分享凭证成功",
		Total:    total,
		Current:  req.Current,
		PageSize: req.PageSize,
		Data:     items,
		Success:  true,
	}, nil
}
