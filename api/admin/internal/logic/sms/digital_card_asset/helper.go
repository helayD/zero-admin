package digital_card_asset

import (
	"context"

	admincommon "github.com/feihua/zero-admin/api/admin/internal/common"
	"github.com/feihua/zero-admin/api/admin/internal/common/errorx"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/feihua/zero-admin/pkg/digitalcardmint"
	pkgscope "github.com/feihua/zero-admin/pkg/scope"
)

func resolveDigitalCardAssetWriteScope(ctx context.Context, requested admincommon.RequestedGovernanceScope) (pkgscope.GovernanceScope, error) {
	current, err := admincommon.ResolveWriteGovernanceScope(ctx, requested)
	if err != nil {
		return pkgscope.GovernanceScope{}, errorx.NewDefaultError(err.Error())
	}
	return current, nil
}

func mapAuditItem(item *digitalcardmint.DigitalCardAssetAuditItem) *types.DigitalCardAssetItem {
	if item == nil {
		return &types.DigitalCardAssetItem{}
	}
	return &types.DigitalCardAssetItem{
		AssetInstanceId:       item.AssetInstanceID,
		AssetNo:               item.AssetNo,
		MemberId:              item.MemberID,
		ActivityId:            item.ActivityID,
		ActivityName:          item.ActivityName,
		TemplateId:            item.TemplateID,
		TemplateName:          item.TemplateName,
		CardFaceImage:         item.CardFaceImage,
		Rarity:                item.Rarity,
		TokenId:               item.TokenID,
		MintTaskId:            item.MintTaskID,
		MintStatus:            item.MintStatus,
		MintStatusText:        item.MintStatusText,
		ChainStatus:           item.ChainStatus,
		ChainStatusText:       item.ChainStatusText,
		DisplayStatus:         item.DisplayStatus,
		DisplayStatusText:     item.DisplayStatusText,
		ComplianceStatus:      item.ComplianceStatus,
		ComplianceStatusText:  item.ComplianceStatusText,
		TokenStatusText:       item.TokenStatusText,
		ComplianceRuleSummary: item.ComplianceRuleSummary,
		LastReceiptSummary:    item.LastReceiptSummary,
		ObtainedAt:            item.ObtainedAt,
		DisposedAt:            item.DisposedAt,
		LatestReasonSummary:   item.LatestReasonSummary,
		ChainType:             item.ChainType,
	}
}

func mapAssetLogs(items []digitalcardmint.AssetLogItem) []types.DigitalCardAssetLogItem {
	result := make([]types.DigitalCardAssetLogItem, 0, len(items))
	for _, item := range items {
		result = append(result, types.DigitalCardAssetLogItem{
			OperationType: item.OperationType,
			OperatorType:  item.OperatorType,
			FromStatus:    item.FromStatus,
			ToStatus:      item.ToStatus,
			ReasonText:    item.ReasonText,
			TraceId:       item.TraceID,
			PayloadJson:   item.PayloadJSON,
			CreateTime:    item.CreateTime,
		})
	}
	return result
}
