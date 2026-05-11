package chainclient

import (
	"errors"
	"time"
)

var ErrReceiptNotFound = errors.New("未找到链上回执")

// ClassifiedError 由具体链客户端实现，用于把底层 SDK 错误归一化为
// (code, reason) 二元组，让上游 digitalcardmint.Service 可以在不依赖具体
// 链实现包的前提下做错误分类（例如：duplicate_mint_rejected 走幂等旁路、
// node_unreachable 走重试、contract_revert 走人工复核）。
//
// Story 10.11 / Task 5.7 / 5.8。
type ClassifiedError interface {
	error
	ChainErrorCode() string
	ChainErrorReason() string
}

// ChainErrorOf 提取链错误的 (code, reason)。如果 err 不是 ClassifiedError，
// 返回 ("", "")，调用方可以据此判断是否为已分类的链错误。
func ChainErrorOf(err error) (code, reason string) {
	if err == nil {
		return "", ""
	}
	var ce ClassifiedError
	if errors.As(err, &ce) {
		return ce.ChainErrorCode(), ce.ChainErrorReason()
	}
	return "", ""
}

// 跨实现共享的标准错误码常量。具体链客户端应优先复用这些 code，
// 避免上游写出散落的字符串比较。Story 10.11 落地 FISCO BCOS 3.x 时定义。
const (
	ChainErrCodeNodeUnreachable     = "node_unreachable"
	ChainErrCodeTLSHandshakeFailed  = "tls_handshake_failed"
	ChainErrCodeGasInsufficient     = "gas_insufficient"
	ChainErrCodeNonceConflict       = "nonce_conflict"
	ChainErrCodeDuplicateMintReject = "duplicate_mint_rejected"
	ChainErrCodeContractRevert      = "contract_revert"
	ChainErrCodeTxTimeout           = "tx_timeout"
	ChainErrCodeReceiptNotFound     = "receipt_not_found"
	ChainErrCodeConfigInvalid       = "config_invalid"
	ChainErrCodeABIDecodeFailed     = "abi_decode_failed"
	ChainErrCodeUnknown             = "unknown"
)

type MintTokenRequest struct {
	IdempotencyKey  string
	TaskID          int64
	AssetInstanceID int64
	ActivityID      int64
	TemplateID      int64
	MemberID        int64
	RequestID       string
	TraceID         string
	AssetNo         string
	ScopeType       string
	PlatformID      int64
	TenantID        int64
	MerchantID      int64
}

type QueryMintTokenRequest struct {
	IdempotencyKey  string
	TaskID          int64
	AssetInstanceID int64
	RequestID       string
	TraceID         string
}

type MintTokenResponse struct {
	TokenID        string
	ChainTxID      string
	ChainStatus    string
	ReceiptSummary string
	ReceiptJSON    string
	ConfirmedAt    time.Time
}
