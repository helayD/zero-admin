// Story 10.11 / Task 5.8 + 8.1：链层错误归一化 + 写库前脱敏。
//
// 设计目标：
//   - 上游不依赖任何具体链实现包（fisco / antchain），仅通过 chainclient.ChainErrorOf
//     提取链层 (code, reason) 用于日志和数据库 last_error_reason 的子码前缀。
//   - 永不把 PEM 证书内容、64+ 字节 hex（疑似私钥）写入数据库，避免
//     运维查日志时把敏感凭据泄露到工单 / 截屏。
//
// 设计选择：
//   - last_error_code 保持 ErrorCodeMintExecuteFailed 父类，避免破坏既有
//     运维报表和告警规则；细分子码以 [code] 前缀方式带在 reason 里。
//   - 当链层未归类（普通 Go error）时，保持原始消息（仍走脱敏）。

package digitalcardmint

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/feihua/zero-admin/pkg/chainclient"
)

// 64 字节 hex 私钥（FISCO BCOS / ECDSA secp256k1 私钥长度），允许可选 0x 前缀。
// 命中即整段替换为 [REDACTED_KEY]。
var hexPrivateKeyPattern = regexp.MustCompile(`(?i)\b(0x)?[0-9a-f]{64}\b`)

// PEM 证书 / 私钥块（含 BEGIN / END）。命中即整段替换为 [REDACTED_PEM]。
var pemBlockPattern = regexp.MustCompile(`(?s)-----BEGIN [A-Z ]+-----.*?-----END [A-Z ]+-----`)

// 4 字节及以上 hex 字符流可能是 tx hash / token id，这些信息要保留（运维定位用），
// 因此只对 64 字符长 hex 做模糊脱敏，不做更激进的处理。

// buildClassifiedFailureReason 根据底层 error 生成可写入 last_error_reason 的脱敏摘要。
//
//   - 如果 err 实现 chainclient.ClassifiedError，前缀 [code] + reason；底层原始消息追加在末尾。
//   - 否则保持原始消息。
//   - 最终统一脱敏 PEM 块和 64 字节 hex 私钥。
func buildClassifiedFailureReason(err error) string {
	if err == nil {
		return ""
	}
	code, reason := chainclient.ChainErrorOf(err)
	raw := err.Error()
	if code != "" {
		// reason 已经是中文人类可读摘要；raw 可能含底层 SDK 文案，便于运维定位。
		// 控制总长度避免无界膨胀（数据库 last_error_reason 通常 VARCHAR(2048)）。
		merged := fmt.Sprintf("[%s] %s | %s", code, reason, raw)
		return sanitizeErrorReason(merged)
	}
	return sanitizeErrorReason(raw)
}

// sanitizeErrorReason 把可能携带 PEM 证书 / 64 字节 hex 私钥的字符串脱敏。
// 同时收敛多余空白字符避免日志被撑爆。
func sanitizeErrorReason(s string) string {
	if s == "" {
		return ""
	}
	out := pemBlockPattern.ReplaceAllString(s, "[REDACTED_PEM]")
	out = hexPrivateKeyPattern.ReplaceAllString(out, "[REDACTED_KEY]")
	out = strings.Join(strings.Fields(out), " ")
	const maxLen = 1024
	if len(out) > maxLen {
		out = out[:maxLen] + "...[truncated]"
	}
	return out
}
