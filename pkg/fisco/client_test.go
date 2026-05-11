// Story 10.11 / Task 5.8: 单元测试覆盖纯逻辑分支。
//
// 不测试需要真实 FISCO 节点的路径（那是 Task 10 的集成测试）。
// 这里覆盖：
//   - disabledClient 错误文案（10.11 Story 明确不要 "免费链" 字样）
//   - invalidConfigClient 校验失败路径
//   - Config.Validate 各种缺字段场景
//   - classifyFiscoError 错误码归一化
//   - decodeRevertReason EVM Error(string) 解码
//   - parseTokenIDFromLogs 模拟 receipt log 解析

package fisco

import (
	"context"
	"encoding/hex"
	"errors"
	"strings"
	"testing"

	bcostypes "github.com/FISCO-BCOS/go-sdk/v3/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/feihua/zero-admin/pkg/chainclient"
)

// ------------------- disabledClient -------------------

func TestDisabledClient_ErrorText(t *testing.T) {
	cli := NewClient(Config{Enabled: false})

	if _, err := cli.MintToken(context.Background(), &chainclient.MintTokenRequest{IdempotencyKey: "k"}); err == nil ||
		!strings.Contains(err.Error(), "FISCO BCOS 客户端未配置") {
		t.Fatalf("MintToken: 期望 'FISCO BCOS 客户端未配置'，实际 %v", err)
	}
	if _, err := cli.QueryMintToken(context.Background(), &chainclient.QueryMintTokenRequest{IdempotencyKey: "k"}); err == nil ||
		!strings.Contains(err.Error(), "FISCO BCOS 客户端未配置") {
		t.Fatalf("QueryMintToken: 期望 'FISCO BCOS 客户端未配置'，实际 %v", err)
	}
	if got := cli.ChainType(); got != ChainTypeFisco3x {
		t.Fatalf("ChainType=%q 应该是 %q（不能再有 free_chain）", got, ChainTypeFisco3x)
	}
	// 旧版含 "免费链" 字样必须不在
	if _, err := cli.MintToken(context.Background(), &chainclient.MintTokenRequest{IdempotencyKey: "k"}); err != nil &&
		strings.Contains(err.Error(), "免费链") {
		t.Fatalf("disabledClient 错误信息不应含 '免费链'：%v", err)
	}
}

// ------------------- invalidConfigClient -------------------

func TestInvalidConfig_HostMissing(t *testing.T) {
	cli := NewClient(Config{
		Enabled:      true,
		ContractAddr: "0xabc",
		PrivateKey:   "0x" + strings.Repeat("11", 32),
		ContractABI:  "[]",
		DisableSsl:   true,
	})
	_, err := cli.MintToken(context.Background(), &chainclient.MintTokenRequest{IdempotencyKey: "k"})
	if err == nil {
		t.Fatal("应该返回 invalid config 错误")
	}
	if !strings.Contains(err.Error(), "Host/Port") {
		t.Fatalf("应包含 Host/Port 错误，实际 %v", err)
	}
	var fe *FiscoError
	if !errors.As(err, &fe) || fe.Code != ErrCodeConfigInvalid {
		t.Fatalf("应为 FiscoError code=%s，实际 %v", ErrCodeConfigInvalid, err)
	}
}

func TestInvalidConfig_TLSPathsMissing(t *testing.T) {
	cli := NewClient(Config{
		Enabled:      true,
		Host:         "127.0.0.1",
		Port:         20200,
		ContractAddr: "0xabc",
		PrivateKey:   "0x" + strings.Repeat("22", 32),
		ContractABI:  "[]",
		// DisableSsl 默认 false，TLS 三个路径必填但都没给
	})
	_, err := cli.QueryMintToken(context.Background(), &chainclient.QueryMintTokenRequest{IdempotencyKey: "k"})
	if err == nil || !strings.Contains(err.Error(), "TLS") {
		t.Fatalf("应包含 TLS 必填错误，实际 %v", err)
	}
}

func TestInvalidConfig_PrivateKeyHexBadFormat(t *testing.T) {
	cli := NewClient(Config{
		Enabled:      true,
		Host:         "127.0.0.1",
		Port:         20200,
		DisableSsl:   true,
		ContractAddr: "0xabc",
		PrivateKey:   "not-a-hex",
		ContractABI:  "[]",
	})
	_, err := cli.MintToken(context.Background(), &chainclient.MintTokenRequest{IdempotencyKey: "k"})
	if err == nil || !strings.Contains(err.Error(), "私钥") {
		t.Fatalf("应包含私钥解析错误，实际 %v", err)
	}
}

// ------------------- Config.Validate -------------------

func TestConfigValidate_DisabledNoCheck(t *testing.T) {
	if err := (Config{Enabled: false}).Validate(); err != nil {
		t.Fatalf("Enabled=false 不应该报错: %v", err)
	}
}

func TestConfigValidate_AllFieldsRequired(t *testing.T) {
	cases := map[string]Config{
		"host_missing": {Enabled: true, Port: 20200, ContractAddr: "0x1", PrivateKey: "p", ContractABI: "[]", DisableSsl: true},
		"port_zero":    {Enabled: true, Host: "127.0.0.1", ContractAddr: "0x1", PrivateKey: "p", ContractABI: "[]", DisableSsl: true},
		"contract_addr_missing": {Enabled: true, Host: "127.0.0.1", Port: 20200, PrivateKey: "p", ContractABI: "[]", DisableSsl: true},
		"private_key_missing": {Enabled: true, Host: "127.0.0.1", Port: 20200, ContractAddr: "0x1", ContractABI: "[]", DisableSsl: true},
		"abi_missing": {Enabled: true, Host: "127.0.0.1", Port: 20200, ContractAddr: "0x1", PrivateKey: "p", DisableSsl: true},
	}
	for name, cfg := range cases {
		t.Run(name, func(t *testing.T) {
			if err := cfg.Validate(); err == nil {
				t.Fatalf("应该报错但 Validate 通过了")
			}
		})
	}
}

func TestConfigValidate_OK(t *testing.T) {
	cfg := Config{
		Enabled:      true,
		Host:         "127.0.0.1",
		Port:         20200,
		DisableSsl:   true, // 跳过 TLS 检查
		ContractAddr: "0x6849f21d1e455e9f0712b1e99fa4fcd23758e8f1",
		ContractABI:  "[]",
		PrivateKey:   "0x" + strings.Repeat("aa", 32),
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("应通过校验: %v", err)
	}
}

func TestConfigDefaults(t *testing.T) {
	cfg := Config{}
	if cfg.resolveGroupID() != DefaultGroupID {
		t.Fatalf("默认 GroupID 应为 %s", DefaultGroupID)
	}
	if cfg.resolveChainID() != DefaultChainID {
		t.Fatalf("默认 ChainID 应为 %s", DefaultChainID)
	}
	if cfg.timeout() != DefaultTimeout {
		t.Fatalf("默认 timeout 应为 %v", DefaultTimeout)
	}
	if cfg.pollIntervalDur() != DefaultPollInterval {
		t.Fatalf("默认 PollInterval 应为 %v", DefaultPollInterval)
	}
	if cfg.pollMaxAttemptsValue() != DefaultPollMaxAttempts {
		t.Fatalf("默认 PollMaxAttempts 应为 %d", DefaultPollMaxAttempts)
	}
}

// ------------------- classifyFiscoError -------------------

func TestClassifyFiscoError_DuplicateMint(t *testing.T) {
	code, _ := classifyFiscoError(errors.New("execution reverted: duplicate mint"))
	if code != ErrCodeDuplicateMintReject {
		t.Fatalf("expect %s, got %s", ErrCodeDuplicateMintReject, code)
	}
}

func TestClassifyFiscoError_NodeUnreachable(t *testing.T) {
	for _, msg := range []string{
		"dial tcp 127.0.0.1:20200: connect: connection refused",
		"read tcp: i/o timeout",
		"network is unreachable",
	} {
		t.Run(msg, func(t *testing.T) {
			code, _ := classifyFiscoError(errors.New(msg))
			if code != ErrCodeNodeUnreachable {
				t.Fatalf("expect %s, got %s for %q", ErrCodeNodeUnreachable, code, msg)
			}
		})
	}
}

func TestClassifyFiscoError_TLS(t *testing.T) {
	code, _ := classifyFiscoError(errors.New("tls: handshake failure: server cert invalid"))
	if code != ErrCodeTLSHandshakeFailed {
		t.Fatalf("expect %s, got %s", ErrCodeTLSHandshakeFailed, code)
	}
}

func TestClassifyFiscoError_Gas(t *testing.T) {
	code, _ := classifyFiscoError(errors.New("gas required exceeds allowance: 21000"))
	if code != ErrCodeGasInsufficient {
		t.Fatalf("expect %s, got %s", ErrCodeGasInsufficient, code)
	}
}

func TestClassifyFiscoError_Nonce(t *testing.T) {
	code, _ := classifyFiscoError(errors.New("nonce too low"))
	if code != ErrCodeNonceConflict {
		t.Fatalf("expect %s, got %s", ErrCodeNonceConflict, code)
	}
}

func TestClassifyFiscoError_GenericRevert(t *testing.T) {
	code, _ := classifyFiscoError(errors.New("execution reverted: not minter"))
	if code != ErrCodeContractRevert {
		t.Fatalf("expect %s, got %s", ErrCodeContractRevert, code)
	}
}

func TestClassifyFiscoError_DuplicateMintTakesPriorityOverGenericRevert(t *testing.T) {
	// 顺序很重要：duplicate mint 必须在 generic revert 前匹配
	code, _ := classifyFiscoError(errors.New("execution reverted: duplicate mint"))
	if code != ErrCodeDuplicateMintReject {
		t.Fatalf("duplicate mint 必须优先于 generic revert，got %s", code)
	}
}

func TestIsDuplicateMint(t *testing.T) {
	dupErr := wrapErr(errors.New("execution reverted: duplicate mint"))
	if !IsDuplicateMint(dupErr) {
		t.Fatal("IsDuplicateMint 应识别")
	}
	otherErr := wrapErr(errors.New("connection refused"))
	if IsDuplicateMint(otherErr) {
		t.Fatal("非 duplicate mint 不应被识别")
	}
}

func TestIsRetriable(t *testing.T) {
	retriable := []string{
		"connection refused",
		"i/o timeout",
		"tls: handshake failure",
		"nonce too low",
		"receipt poll timeout",
	}
	for _, m := range retriable {
		if !IsRetriable(wrapErr(errors.New(m))) {
			t.Fatalf("%q 应为可重试", m)
		}
	}
	notRetriable := []string{
		"execution reverted: duplicate mint",
		"execution reverted: not minter",
		"gas required exceeds allowance",
	}
	for _, m := range notRetriable {
		if IsRetriable(wrapErr(errors.New(m))) {
			t.Fatalf("%q 不应可重试", m)
		}
	}
}

// ------------------- decodeRevertReason -------------------

func TestDecodeRevertReason_DuplicateMint(t *testing.T) {
	// 模拟 EVM Error("duplicate mint") encoding:
	// selector(4B) || offset(32B=0x20) || length(32B=14) || "duplicate mint"+padding
	selector := crypto.Keccak256([]byte("Error(string)"))[:4]
	out := append([]byte{}, selector...)
	// offset = 0x20
	out = append(out, leftPad32(0x20)...)
	msg := []byte("duplicate mint")
	out = append(out, leftPad32(int64(len(msg)))...)
	out = append(out, padRight(msg, 32)...)

	got := decodeRevertReason(hex.EncodeToString(out))
	if got != "duplicate mint" {
		t.Fatalf("期望 'duplicate mint'，实际 %q", got)
	}
}

func TestDecodeRevertReason_Empty(t *testing.T) {
	if decodeRevertReason("") != "" {
		t.Fatal("空 output 应返回空")
	}
}

func TestDecodeRevertReason_BadSelector(t *testing.T) {
	// selector 不是 Error(string) 时返回空
	bogus := strings.Repeat("ff", 4) + strings.Repeat("00", 64)
	if decodeRevertReason(bogus) != "" {
		t.Fatal("非 Error(string) selector 应返回空")
	}
}

// ------------------- parseTokenIDFromLogs -------------------

func TestParseTokenIDFromLogs_Match(t *testing.T) {
	c := &fiscoRealClient{
		cardMintedID: crypto.Keccak256Hash([]byte("CardMinted(address,uint256,string,string)")),
	}
	tokenID := int64(42)
	logs := []*bcostypes.NewLog{
		{
			Topics: []string{
				c.cardMintedID.Hex(),
				common.BytesToHash([]byte{0xCA, 0xFE}).Hex(),
				"0x" + hex.EncodeToString(leftPad32(tokenID)),
			},
			Data:    "0x",
			Address: "0xabc",
		},
	}
	got, err := c.parseTokenIDFromLogs(logs)
	if err != nil {
		t.Fatalf("parseTokenIDFromLogs error: %v", err)
	}
	if got != "42" {
		t.Fatalf("tokenId 应为 42，实际 %s", got)
	}
}

func TestParseTokenIDFromLogs_NoMatch(t *testing.T) {
	c := &fiscoRealClient{
		cardMintedID: crypto.Keccak256Hash([]byte("CardMinted(address,uint256,string,string)")),
	}
	logs := []*bcostypes.NewLog{
		{Topics: []string{"0x" + strings.Repeat("aa", 32), "0xb", "0xc"}},
	}
	if _, err := c.parseTokenIDFromLogs(logs); err == nil {
		t.Fatal("不匹配的 topic 应返回 error")
	}
}

// ------------------- buildMetadataURI -------------------

func TestBuildMetadataURI(t *testing.T) {
	uri, err := buildMetadataURI(&chainclient.MintTokenRequest{
		IdempotencyKey: "card-mint:970001",
		AssetNo:        "ASSET-001",
		ActivityID:     930001,
		TraceID:        "trace-1",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(uri, "data:application/json;charset=utf-8,") {
		t.Fatalf("metadata URI 前缀错误: %s", uri)
	}
	if !strings.Contains(uri, "card-mint:970001") {
		t.Fatalf("metadata URI 应含 idempotencyKey")
	}
	if !strings.Contains(uri, "ASSET-001") {
		t.Fatalf("metadata URI 应含 assetNo")
	}
}

// ------------------- helpers -------------------

func leftPad32(v int64) []byte {
	out := make([]byte, 32)
	for i := 31; i >= 0 && v > 0; i-- {
		out[i] = byte(v & 0xff)
		v >>= 8
	}
	return out
}

func padRight(b []byte, multiple int) []byte {
	r := len(b) % multiple
	if r == 0 {
		return b
	}
	return append(b, make([]byte, multiple-r)...)
}

// ------------------- 接口契约 -------------------

func TestImplementsChainClient(t *testing.T) {
	var _ chainclient.ChainClient = (*fiscoRealClient)(nil)
	var _ chainclient.ChainClient = (disabledClient{})
	var _ chainclient.ChainClient = (*invalidConfigClient)(nil)
}
