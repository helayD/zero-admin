package digitalcardmint

import (
	"errors"
	"strings"
	"testing"

	"github.com/feihua/zero-admin/pkg/chainclient"
)

type fakeClassifiedError struct {
	code   string
	reason string
}

func (e *fakeClassifiedError) Error() string            { return "raw sdk: " + e.reason }
func (e *fakeClassifiedError) ChainErrorCode() string   { return e.code }
func (e *fakeClassifiedError) ChainErrorReason() string { return e.reason }

func TestBuildClassifiedFailureReason_WithClassifiedError(t *testing.T) {
	err := &fakeClassifiedError{code: chainclient.ChainErrCodeNodeUnreachable, reason: "FISCO 节点不可达"}
	got := buildClassifiedFailureReason(err)
	if !strings.HasPrefix(got, "[node_unreachable] ") {
		t.Fatalf("expected prefix [node_unreachable] in %q", got)
	}
	if !strings.Contains(got, "FISCO 节点不可达") {
		t.Fatalf("expected reason in %q", got)
	}
}

func TestBuildClassifiedFailureReason_WithPlainError(t *testing.T) {
	got := buildClassifiedFailureReason(errors.New("dial tcp 127.0.0.1:20200: connection refused"))
	if strings.Contains(got, "[") {
		t.Fatalf("plain error must not have classified prefix, got %q", got)
	}
	if !strings.Contains(got, "connection refused") {
		t.Fatalf("plain error must keep raw text, got %q", got)
	}
}

func TestBuildClassifiedFailureReason_NilSafe(t *testing.T) {
	if got := buildClassifiedFailureReason(nil); got != "" {
		t.Fatalf("nil should produce empty, got %q", got)
	}
}

func TestSanitizeErrorReason_RedactsPEMBlock(t *testing.T) {
	pem := "-----BEGIN CERTIFICATE-----\nMIIBxxx\nfoo\n-----END CERTIFICATE-----"
	in := "tls: bad cert: " + pem + " ; trailing"
	got := sanitizeErrorReason(in)
	if strings.Contains(got, "BEGIN CERTIFICATE") {
		t.Fatalf("PEM block should be redacted, got %q", got)
	}
	if !strings.Contains(got, "[REDACTED_PEM]") {
		t.Fatalf("expected [REDACTED_PEM] marker in %q", got)
	}
}

func TestSanitizeErrorReason_RedactsHexPrivateKey(t *testing.T) {
	pk := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef" // 64 hex
	in := "sign error key=" + pk + " end"
	got := sanitizeErrorReason(in)
	if strings.Contains(got, pk) {
		t.Fatalf("64-byte hex should be redacted, got %q", got)
	}
	if !strings.Contains(got, "[REDACTED_KEY]") {
		t.Fatalf("expected [REDACTED_KEY] marker in %q", got)
	}
}

func TestSanitizeErrorReason_PreservesShortHex(t *testing.T) {
	// 短 hex（如 tx hash 64 hex 也算） — 注意：tx hash 也是 64 hex，会被误伤。
	// 这里测试非 64 长度的 hex（区块号、错误码等）不会被吞。
	in := "block=12345 errorCode=0x1f"
	got := sanitizeErrorReason(in)
	if !strings.Contains(got, "block=12345") || !strings.Contains(got, "0x1f") {
		t.Fatalf("non-64 hex must be preserved, got %q", got)
	}
}

func TestSanitizeErrorReason_TrimsWhitespace(t *testing.T) {
	in := "  multiple   spaces\n\nand\ttabs  "
	got := sanitizeErrorReason(in)
	if got != "multiple spaces and tabs" {
		t.Fatalf("expected normalized whitespace, got %q", got)
	}
}

func TestSanitizeErrorReason_TruncatesLongInput(t *testing.T) {
	in := strings.Repeat("a", 2000)
	got := sanitizeErrorReason(in)
	if !strings.HasSuffix(got, "...[truncated]") {
		t.Fatalf("expected truncation marker, got len=%d", len(got))
	}
	if len(got) > 1024+len("...[truncated]") {
		t.Fatalf("truncated length out of bound: %d", len(got))
	}
}

func TestSanitizeErrorReason_EmptyInput(t *testing.T) {
	if got := sanitizeErrorReason(""); got != "" {
		t.Fatalf("empty input should produce empty output, got %q", got)
	}
}
