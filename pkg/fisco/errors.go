// Story 10.11 / Task 5.7: FISCO BCOS 错误归一化。
//
// 上游 Logic 拿到 (code, reason) 后写入 sms_card_mint_task.fail_reason / fail_code，
// 并据此决定重试 / 进死信 / 走 QueryMintToken 旁路。

package fisco

import (
	"errors"
	"strings"
)

// 错误代码常量，供上游做分支判断（例如 duplicate_mint_rejected → fallback to QueryMintToken）。
const (
	ErrCodeNodeUnreachable     = "node_unreachable"
	ErrCodeTLSHandshakeFailed  = "tls_handshake_failed"
	ErrCodeGasInsufficient     = "gas_insufficient"
	ErrCodeNonceConflict       = "nonce_conflict"
	ErrCodeDuplicateMintReject = "duplicate_mint_rejected"
	ErrCodeContractRevert      = "contract_revert"
	ErrCodeTxTimeout           = "tx_timeout"
	ErrCodeReceiptNotFound     = "receipt_not_found"
	ErrCodeConfigInvalid       = "config_invalid"
	ErrCodeABIDecodeFailed     = "abi_decode_failed"
	ErrCodeUnknown             = "unknown"
)

// FiscoError 包装底层错误，附带稳定的错误代码和原因。
type FiscoError struct {
	Code   string
	Reason string
	Err    error
}

func (e *FiscoError) Error() string {
	if e.Err != nil {
		return "fisco[" + e.Code + "]: " + e.Reason + ": " + e.Err.Error()
	}
	return "fisco[" + e.Code + "]: " + e.Reason
}

func (e *FiscoError) Unwrap() error { return e.Err }

// ChainErrorCode / ChainErrorReason 实现 chainclient.ClassifiedError 接口，
// 让 digitalcardmint.Service 在不依赖 fisco 包的前提下读取归一化错误码。
func (e *FiscoError) ChainErrorCode() string   { return e.Code }
func (e *FiscoError) ChainErrorReason() string { return e.Reason }

// IsDuplicateMint 上游用于判断是否走 QueryMintToken 旁路。
func IsDuplicateMint(err error) bool {
	var fe *FiscoError
	if errors.As(err, &fe) {
		return fe.Code == ErrCodeDuplicateMintReject
	}
	return false
}

// IsRetriable 暂时性错误（节点不可达 / TLS / 超时 / nonce）适合重试。
// gas / contract revert 不重试。
func IsRetriable(err error) bool {
	var fe *FiscoError
	if !errors.As(err, &fe) {
		return false
	}
	switch fe.Code {
	case ErrCodeNodeUnreachable,
		ErrCodeTLSHandshakeFailed,
		ErrCodeNonceConflict,
		ErrCodeTxTimeout:
		return true
	default:
		return false
	}
}

// classifyFiscoError 将 SDK 原始错误映射为 (code, reason)。
//
// 匹配按从严到宽顺序：先匹配可能携带 "execution reverted: duplicate mint" 这种最具体的
// 业务幂等信号，避免被 contract_revert 通用桶兜底。
func classifyFiscoError(err error) (code, reason string) {
	if err == nil {
		return "", ""
	}
	msg := strings.ToLower(err.Error())

	// 业务幂等信号（必须最先匹配）
	if strings.Contains(msg, "duplicate mint") {
		return ErrCodeDuplicateMintReject, "合约幂等拒绝（duplicate mint）"
	}

	// 网络层
	if strings.Contains(msg, "connection refused") ||
		strings.Contains(msg, "no route to host") ||
		strings.Contains(msg, "network is unreachable") ||
		strings.Contains(msg, "i/o timeout") ||
		strings.Contains(msg, "broken pipe") ||
		strings.Contains(msg, "connection reset") {
		return ErrCodeNodeUnreachable, "FISCO 节点不可达"
	}

	// TLS
	if strings.Contains(msg, "tls: handshake failure") ||
		strings.Contains(msg, "x509:") ||
		strings.Contains(msg, "certificate") {
		return ErrCodeTLSHandshakeFailed, "TLS 双向认证失败"
	}

	// 交易费 / nonce
	if strings.Contains(msg, "gas required exceeds allowance") ||
		strings.Contains(msg, "insufficient funds") ||
		strings.Contains(msg, "out of gas") {
		return ErrCodeGasInsufficient, "Gas 不足"
	}
	if strings.Contains(msg, "nonce too low") || strings.Contains(msg, "nonce too high") {
		return ErrCodeNonceConflict, "Nonce 冲突"
	}

	// 轮询超时
	if strings.Contains(msg, "receipt poll timeout") || strings.Contains(msg, "tx timeout") {
		return ErrCodeTxTimeout, "交易回执轮询超时"
	}

	// 通用合约 revert（兜底，必须放在 duplicate mint 之后）
	if strings.Contains(msg, "execution reverted") || strings.Contains(msg, "revert") {
		return ErrCodeContractRevert, "合约执行被 revert"
	}

	// ABI / 编码
	if strings.Contains(msg, "abi:") || strings.Contains(msg, "unmarshal") {
		return ErrCodeABIDecodeFailed, "ABI 解码失败"
	}

	return ErrCodeUnknown, "未分类错误"
}

// wrapErr 把原始 err 归一化为 *FiscoError。nil 透传。
func wrapErr(err error) error {
	if err == nil {
		return nil
	}
	code, reason := classifyFiscoError(err)
	return &FiscoError{Code: code, Reason: reason, Err: err}
}
