package digital_card_asset

import (
	"context"

	frontcommon "github.com/feihua/zero-admin/api/front/internal/logic/common"
	"github.com/feihua/zero-admin/api/front/internal/types"
	"github.com/feihua/zero-admin/pkg/digitalcardmint"
	"github.com/feihua/zero-admin/pkg/errorx"
	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	"github.com/zeromicro/go-zero/core/logc"
)

const assetCodeSuccess = "SUCCESS"

func currentMemberID(ctx context.Context) (int64, error) {
	return frontcommon.GetMemberId(ctx)
}

func currentGovernanceScope(ctx context.Context) pkgscope.GovernanceScope {
	return frontcommon.ResolveEffectiveGovernanceScope(ctx)
}

func assetServiceError(ctx context.Context, action string, payload interface{}, err error) error {
	logc.Errorf(ctx, "%s失败,参数:%+v,异常:%s", action, payload, err.Error())
	return errorx.NewDefaultError(err.Error())
}

func mapAssetItem(item digitalcardmint.MemberDigitalCardAssetItem) types.DigitalCardAssetItem {
	return types.DigitalCardAssetItem{
		AssetInstanceId:       item.AssetInstanceID,
		AssetNo:               item.AssetNo,
		TemplateId:            item.TemplateID,
		TemplateName:          item.TemplateName,
		CardFaceImage:         item.CardFaceImage,
		ActivityId:            item.ActivityID,
		ActivityName:          item.ActivityName,
		Rarity:                item.Rarity,
		ObtainedAt:            item.ObtainedAt,
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
		ChainType:             item.ChainType,
	}
}

func mapTimeline(items []digitalcardmint.MemberDigitalCardAssetTimelineItem) []types.DigitalCardAssetTimelineItem {
	result := make([]types.DigitalCardAssetTimelineItem, 0, len(items))
	for _, item := range items {
		result = append(result, types.DigitalCardAssetTimelineItem{
			OperationType: item.OperationType,
			OperationText: item.OperationText,
			StatusText:    item.StatusText,
			ReasonText:    item.ReasonText,
			CreateTime:    item.CreateTime,
		})
	}
	return result
}

func mapDrawSummary(item digitalcardmint.MemberDigitalCardAssetDrawSummary) types.DigitalCardAssetDrawSummary {
	return types.DigitalCardAssetDrawSummary{
		ParticipationRecordId: item.ParticipationRecordID,
		ResultType:            item.ResultType,
		ResultStatus:          item.ResultStatus,
		ResultStatusText:      item.ResultStatusText,
		FailureReason:         item.FailureReason,
		CreateTime:            item.CreateTime,
	}
}
