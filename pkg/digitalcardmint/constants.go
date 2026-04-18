package digitalcardmint

import "strings"

const (
	AssetStatusCreated = "asset_created"

	TaskStatusPendingDispatch = "pending_dispatch"
	TaskStatusDispatched      = "dispatched"
	TaskStatusRunning         = "running"
	TaskStatusSucceeded       = "succeeded"
	TaskStatusFailed          = "failed"
	TaskStatusManualReview    = "manual_review"
	TaskStatusFrozen          = "frozen"

	MintStatusPending      = "mint_pending"
	MintStatusProcessing   = "mint_processing"
	MintStatusSuccess      = "mint_success"
	MintStatusFailed       = "mint_failed"
	MintStatusCompensating = "mint_compensating"
	MintStatusManualReview = "mint_manual_review"
	MintStatusFrozen       = "mint_frozen"

	ChainStatusUnknown    = "unknown"
	ChainStatusProcessing = "processing"
	ChainStatusSuccess    = "success"
	ChainStatusFailed     = "failed"
	ChainStatusFrozen     = "frozen"

	OperationMintRequested      = "mint_requested"
	OperationMintDispatching    = "mint_dispatching"
	OperationMintSucceeded      = "mint_succeeded"
	OperationMintFailed         = "mint_failed"
	OperationMintRetryRequested = "mint_retry_requested"
	OperationMintFrozen         = "mint_frozen"
	OperationMintManualReview   = "mint_manual_review"

	OperatorSystem = "system"
	OperatorJob    = "job"
	OperatorManual = "manual"

	EventNameMintRequested = "sms.digital_card.mint_requested.v1"
	EventExchange          = "sms.event.exchange"
	EventExchangeType      = "topic"
	EventQueue             = "sms.digital_card.mint.queue"
	EventRoutingKey        = "sms.digital_card.mint_requested.key"

	DefaultMaxRetryCount int32 = 3

	ErrorCodeMQDispatchFailed         = "mq_dispatch_failed"
	ErrorCodeMintExecuteFailed        = "mint_execute_failed"
	ErrorCodeReceiptWritebackFailed   = "receipt_writeback_failed"
	ErrorCodeReceiptReconcileRequired = "receipt_reconcile_required"
	ErrorCodeMintPrerequisiteRejected = "mint_prerequisite_rejected"
	ErrorCodeTokenBindingConflict     = "token_binding_conflict"
)

func ResolveAssetStatusText(assetStatus, mintStatus, chainStatus string) string {
	switch strings.TrimSpace(mintStatus) {
	case MintStatusSuccess:
		return "链上确权成功"
	case MintStatusCompensating:
		return "链上补偿处理中"
	case MintStatusManualReview:
		return "链路升级为人工复核"
	case MintStatusFrozen:
		return "链路已冻结"
	case MintStatusFailed:
		return "链上发放失败，等待补偿"
	case MintStatusProcessing:
		return "资产已创建，链上处理中"
	}

	switch strings.TrimSpace(chainStatus) {
	case ChainStatusSuccess:
		return "链上确权成功"
	case ChainStatusFrozen:
		return "链路已冻结"
	case ChainStatusFailed:
		return "链上发放失败，等待补偿"
	}

	if strings.TrimSpace(assetStatus) == AssetStatusCreated {
		return "资产已创建，链上处理中"
	}
	return "资产处理中"
}
