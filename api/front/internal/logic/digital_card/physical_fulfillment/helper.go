package physical_fulfillment

import (
	"context"

	frontcommon "github.com/feihua/zero-admin/api/front/internal/logic/common"
	"github.com/feihua/zero-admin/api/front/internal/types"
	"github.com/feihua/zero-admin/pkg/digitalcardmint"
	"github.com/feihua/zero-admin/pkg/errorx"
	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	"github.com/zeromicro/go-zero/core/logc"
)

const physicalCodeSuccess = "SUCCESS"

func currentMemberID(ctx context.Context) (int64, error) {
	return frontcommon.GetMemberId(ctx)
}

func currentGovernanceScope(ctx context.Context) pkgscope.GovernanceScope {
	return frontcommon.ResolveEffectiveGovernanceScope(ctx)
}

func physicalServiceError(ctx context.Context, action string, payload interface{}, err error) error {
	logc.Errorf(ctx, "%s失败,参数:%+v,异常:%s", action, payload, err.Error())
	return errorx.NewDefaultError(err.Error())
}

func mapPhysicalDetail(detail *digitalcardmint.MemberPhysicalFulfillmentDetail, blocked *digitalcardmint.PhysicalFulfillmentResult, assetInstanceID int64) types.PhysicalFulfillmentDetailData {
	if blocked != nil && blocked.BlockedReason != "" {
		return types.PhysicalFulfillmentDetailData{
			AssetInstanceId:       assetInstanceID,
			FulfillmentStatus:     blocked.FulfillmentStatus,
			FulfillmentStatusText: blocked.FulfillmentStatusText,
			ShippingFeeStatus:     blocked.ShippingFeeStatus,
			ShippingFeeStatusText: blocked.ShippingFeeStatusText,
			ShippingFeeAmount:     blocked.ShippingFeeAmount,
			BlockedReason:         blocked.BlockedReason,
			BlockedReasonText:     blocked.BlockedReasonText,
			Timeline:              []types.PhysicalFulfillmentTimelineItem{},
		}
	}
	if detail == nil {
		return types.PhysicalFulfillmentDetailData{
			AssetInstanceId: assetInstanceID,
			Timeline:        []types.PhysicalFulfillmentTimelineItem{},
		}
	}
	return types.PhysicalFulfillmentDetailData{
		FulfillmentId:         detail.FulfillmentID,
		FulfillmentNo:         detail.FulfillmentNo,
		AssetInstanceId:       detail.AssetInstanceID,
		AssetNo:               detail.AssetNo,
		TemplateName:          detail.TemplateName,
		ActivityName:          detail.ActivityName,
		ObtainedAt:            detail.ObtainedAt,
		MintStatusText:        detail.MintStatusText,
		FulfillmentStatus:     detail.FulfillmentStatus,
		FulfillmentStatusText: detail.FulfillmentStatusText,
		ProductionStatusText:  detail.ProductionStatusText,
		ShippingStatusText:    detail.ShippingStatusText,
		ShippingFeeStatus:     detail.ShippingFeeStatus,
		ShippingFeeStatusText: detail.ShippingFeeStatusText,
		ShippingFeeAmount:     detail.ShippingFeeAmount,
		ReceiverNameMasked:    detail.ReceiverNameMasked,
		ReceiverPhoneMasked:   detail.ReceiverPhoneMasked,
		AddressSummary:        detail.AddressSummary,
		CarrierName:           detail.CarrierName,
		TrackingNo:            detail.TrackingNo,
		ComplianceTipSummary:  detail.ComplianceTipSummary,
		Timeline:              mapPhysicalTimeline(detail.Timeline),
	}
}

func mapPhysicalTimeline(items []digitalcardmint.PhysicalFulfillmentTimelineItem) []types.PhysicalFulfillmentTimelineItem {
	result := make([]types.PhysicalFulfillmentTimelineItem, 0, len(items))
	for _, item := range items {
		result = append(result, types.PhysicalFulfillmentTimelineItem{
			Action:     item.Action,
			ActionText: item.ActionText,
			StatusText: item.StatusText,
			Reason:     item.Reason,
			CreateTime: item.CreateTime,
		})
	}
	return result
}
