package digital_card_physical_fulfillment

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	admincommon "github.com/feihua/zero-admin/api/admin/internal/common"
	"github.com/feihua/zero-admin/api/admin/internal/common/errorx"
	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/feihua/zero-admin/api/admin/internal/types"
	"github.com/feihua/zero-admin/pkg/digitalcardmint"
	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	"github.com/feihua/zero-admin/rpc/sys/sysclient"
)

const digitalCardPhysicalFulfillmentBusinessType int32 = 2

func resolvePhysicalFulfillmentWriteScope(ctx context.Context, requested admincommon.RequestedGovernanceScope) (pkgscope.GovernanceScope, error) {
	current, err := admincommon.ResolveWriteGovernanceScope(ctx, requested)
	if err != nil {
		return pkgscope.GovernanceScope{}, errorx.NewDefaultError(err.Error())
	}
	return current, nil
}

func resolvePhysicalActionContext(ctx context.Context, scopeType string, platformID int64, tenantID int64, merchantID int64) (pkgscope.GovernanceScope, int64, error) {
	writeScope, err := resolvePhysicalFulfillmentWriteScope(ctx, requestedPhysicalScope(scopeType, platformID, tenantID, merchantID))
	if err != nil {
		return pkgscope.GovernanceScope{}, 0, err
	}
	operatorID, err := admincommon.GetUserId(ctx)
	if err != nil {
		return pkgscope.GovernanceScope{}, 0, errorx.NewDefaultError("无法获取操作人身份，请重新登录")
	}
	return writeScope, operatorID, nil
}

func requestedPhysicalScope(scopeType string, platformID int64, tenantID int64, merchantID int64) admincommon.RequestedGovernanceScope {
	return admincommon.RequestedGovernanceScope{
		ScopeType:  scopeType,
		PlatformID: platformID,
		TenantID:   tenantID,
		MerchantID: merchantID,
	}
}

func mapPhysicalFulfillmentItem(item *digitalcardmint.PhysicalFulfillmentItem) *types.DigitalCardPhysicalFulfillmentItem {
	if item == nil {
		return &types.DigitalCardPhysicalFulfillmentItem{}
	}
	return &types.DigitalCardPhysicalFulfillmentItem{
		FulfillmentId:         item.FulfillmentID,
		FulfillmentNo:         item.FulfillmentNo,
		AssetInstanceId:       item.AssetInstanceID,
		AssetNo:               item.AssetNo,
		ActivityId:            item.ActivityID,
		ActivityName:          item.ActivityName,
		TemplateId:            item.TemplateID,
		TemplateName:          item.TemplateName,
		MemberId:              item.MemberID,
		FulfillmentStatus:     item.FulfillmentStatus,
		FulfillmentStatusText: item.FulfillmentStatusText,
		ProductionStatus:      item.ProductionStatus,
		ProductionStatusText:  item.ProductionStatusText,
		ShippingStatus:        item.ShippingStatus,
		ShippingStatusText:    item.ShippingStatusText,
		ProductionBatchNo:     item.ProductionBatchNo,
		CarrierName:           item.CarrierName,
		TrackingNo:            item.TrackingNo,
		FailureCode:           item.FailureCode,
		FailureReason:         item.FailureReason,
		ShippedAt:             item.ShippedAt,
		SignedAt:              item.SignedAt,
		UpdateTime:            item.UpdateTime,
	}
}

func mapPhysicalTimeline(items []digitalcardmint.PhysicalFulfillmentTimelineItem) []types.DigitalCardPhysicalFulfillmentTimelineItem {
	result := make([]types.DigitalCardPhysicalFulfillmentTimelineItem, 0, len(items))
	for _, item := range items {
		result = append(result, types.DigitalCardPhysicalFulfillmentTimelineItem{
			Action:     item.Action,
			ActionText: item.ActionText,
			StatusText: item.StatusText,
			Reason:     item.Reason,
			CreateTime: item.CreateTime,
		})
	}
	return result
}

func mapPhysicalActionResp(message string, result *digitalcardmint.PhysicalFulfillmentResult) *types.DigitalCardPhysicalFulfillmentActionResp {
	if result == nil {
		result = &digitalcardmint.PhysicalFulfillmentResult{}
	}
	return &types.DigitalCardPhysicalFulfillmentActionResp{
		Code:                  "000000",
		Message:               message,
		FulfillmentId:         result.FulfillmentID,
		AssetInstanceId:       result.AssetInstanceID,
		FulfillmentNo:         result.FulfillmentNo,
		FulfillmentStatus:     result.FulfillmentStatus,
		FulfillmentStatusText: result.FulfillmentStatusText,
		ProductionStatus:      result.ProductionStatus,
		ProductionStatusText:  result.ProductionStatusText,
		ShippingStatus:        result.ShippingStatus,
		ShippingStatusText:    result.ShippingStatusText,
		BlockedReason:         result.BlockedReason,
		BlockedReasonText:     result.BlockedReasonText,
		Success:               true,
	}
}

func writePhysicalFulfillmentOperateLog(ctx context.Context, svcCtx *svc.ServiceContext, operatorID int64, action string, fulfillmentID int64, reason string, result *digitalcardmint.PhysicalFulfillmentResult) {
	if operatorID <= 0 || svcCtx == nil || svcCtx.Operatelogservice == nil || result == nil {
		return
	}
	current, _ := admincommon.CurrentGovernanceScope(ctx)
	operateParam := marshalPhysicalOperateLogJSON(map[string]interface{}{
		"action":        action,
		"fulfillmentId": fulfillmentID,
		"reason":        reason,
	})
	jsonResult := marshalPhysicalOperateLogJSON(map[string]interface{}{
		"fulfillmentId":     result.FulfillmentID,
		"fulfillmentStatus": result.FulfillmentStatus,
		"productionStatus":  result.ProductionStatus,
		"shippingStatus":    result.ShippingStatus,
		"blockedReason":     result.BlockedReason,
	})
	extra := marshalPhysicalOperateLogJSON(map[string]interface{}{
		"action":          action,
		"assetInstanceId": result.AssetInstanceID,
		"fulfillmentNo":   result.FulfillmentNo,
	})
	_, _ = svcCtx.Operatelogservice.AddOperateLog(ctx, &sysclient.AddOperateLogReq{
		Title:         "实体卡履约-" + action,
		BusinessType:  digitalCardPhysicalFulfillmentBusinessType,
		Method:        "/api/sms/digitalCardPhysicalFulfillment/" + action,
		RequestMethod: "POST",
		OperatorType:  1,
		OperateUrl:    "/api/sms/digitalCardPhysicalFulfillment/" + action,
		Platform:      "admin",
		Status:        0,
		OperateTime:   time.Now().Format("2006-01-02 15:04:05"),
		OperateName:   strconv.FormatInt(operatorID, 10),
		DeptName:      strconv.FormatInt(current.TenantID, 10),
		OperateParam:  operateParam,
		JsonResult:    jsonResult,
		Extra:         extra,
	})
}

func marshalPhysicalOperateLogJSON(payload interface{}) string {
	body, err := json.Marshal(payload)
	if err != nil {
		return "{}"
	}
	return string(body)
}
