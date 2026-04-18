package digitalcardmint

import "strings"

const (
	DisplayStatusVisible  = "display_visible"
	DisplayStatusHidden   = "display_hidden"
	DisplayStatusOfflined = "display_offlined"
	DisplayStatusRecycled = "display_recycled"

	ComplianceStatusClear            = "compliance_clear"
	ComplianceStatusReview           = "compliance_review"
	ComplianceStatusRestricted       = "compliance_restricted"
	ComplianceStatusRecycleRequested = "compliance_recycle_requested"
	ComplianceStatusRecycled         = "compliance_recycled"

	OperationAssetDisplayHidden    = "asset_display_hidden"
	OperationAssetDisplayRestored  = "asset_display_restored"
	OperationAssetComplianceReview = "asset_compliance_review"
	OperationAssetOfflined         = "asset_offlined"
	OperationAssetRecycleRequested = "asset_recycle_requested"
	OperationAssetRecycled         = "asset_recycled"
)

func displayStatusText(status string) string {
	switch strings.TrimSpace(status) {
	case DisplayStatusVisible:
		return "正常展示"
	case DisplayStatusHidden:
		return "受限展示"
	case DisplayStatusOfflined:
		return "已下线展示"
	case DisplayStatusRecycled:
		return "已回收"
	default:
		return "展示状态未知"
	}
}

func complianceStatusText(status string) string {
	switch strings.TrimSpace(status) {
	case ComplianceStatusClear:
		return "合规正常"
	case ComplianceStatusReview:
		return "人工复核中"
	case ComplianceStatusRestricted:
		return "已限制展示"
	case ComplianceStatusRecycleRequested:
		return "回收处理中"
	case ComplianceStatusRecycled:
		return "已回收"
	default:
		return "合规状态未知"
	}
}

func tokenStatusText(tokenID string, chainStatus string) string {
	if strings.TrimSpace(tokenID) != "" && strings.TrimSpace(chainStatus) == ChainStatusSuccess {
		return "已确认 token"
	}
	return chainStatusText(chainStatus)
}

func complianceRuleSummary(activitySummary string, displayReason string, complianceReason string) string {
	return firstNonEmpty(
		strings.TrimSpace(complianceReason),
		strings.TrimSpace(displayReason),
		strings.TrimSpace(activitySummary),
	)
}

func assetLogOperationText(operationType string) string {
	switch strings.TrimSpace(operationType) {
	case OperationMintRequested:
		return "已创建发放任务"
	case OperationMintDispatching:
		return "已派发链路任务"
	case OperationMintSucceeded:
		return "链上发放成功"
	case OperationMintFailed:
		return "链路进入补偿"
	case OperationMintRetryRequested:
		return "人工触发重试"
	case OperationMintFrozen:
		return "链路已冻结"
	case OperationMintManualReview:
		return "链路升级人工复核"
	case OperationAssetDisplayHidden:
		return "资产改为受限展示"
	case OperationAssetDisplayRestored:
		return "资产恢复展示"
	case OperationAssetComplianceReview:
		return "标记资产人工复核"
	case OperationAssetOfflined:
		return "资产已下线展示"
	case OperationAssetRecycleRequested:
		return "发起回收处置"
	case OperationAssetRecycled:
		return "资产已回收"
	default:
		return strings.TrimSpace(operationType)
	}
}

func maskTokenID(tokenID string) string {
	value := strings.TrimSpace(tokenID)
	if value == "" {
		return ""
	}
	if len(value) <= 8 {
		return value[:2] + "****"
	}
	return value[:4] + "****" + value[len(value)-4:]
}
