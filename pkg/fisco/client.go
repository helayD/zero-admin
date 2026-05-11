// Story 10.11 / Task 5: 真实 FISCO BCOS 3.x 客户端。
//
// 实现 ChainClient 接口，与 antchain.NewClient 对称。
//
// 关键流程（MintToken）：
//  1. 私钥派生 from-address
//  2. metadata JSON 组装（assetNo / activityId / templateId / memberId / scope / traceId）
//  3. 通过合约 ABI Pack mint(to, idempotencyKey, metadataJSON) calldata
//  4. CreateEncodedTransactionDataV1 → CreateEncodedSignature → CreateEncodedTransaction
//  5. SendEncodedTransaction(ctx, tx, true) → 同步返回 *types.Receipt（SDK 内部已轮询）
//  6. receipt.Status != 0 → classifyFiscoError 映射；duplicate mint → ErrCodeDuplicateMintReject
//  7. 从 Receipt.Logs 中匹配 CardMinted event topic[0]，topic[2] 即 uint256 tokenId（十进制字符串）
//  8. 构造 chainclient.MintTokenResponse（TokenID / ChainTxID / ChainStatus / ReceiptSummary / ReceiptJSON / ConfirmedAt）
//
// 关键流程（QueryMintToken）：
//  1. ABI Pack getTokenByKey(idempotencyKey) calldata
//  2. CallContract（read-only）返回 32 字节大端 uint256
//  3. 0 → 返回 chainclient.ErrReceiptNotFound（上游回退到正常 mint 流程）
//  4. 非 0 → 构造 MintTokenResponse（不持久化 tx hash，本 Story 暂用 view-only 数据；
//          下一 Story 加 sms_chain_receipt_cache 表后再补 tx hash 回查）
//
// SDK 连接采取 lazy 模式：NewClient 仅做 Config 校验和 ABI 解析，
// 实际 dial 在第一次 MintToken / QueryMintToken 时执行（避免服务启动时硬依赖链可达）。

package fisco

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"sync"
	"time"

	bcosabi "github.com/FISCO-BCOS/go-sdk/v3/abi"
	bcosclient "github.com/FISCO-BCOS/go-sdk/v3/client"
	bcostypes "github.com/FISCO-BCOS/go-sdk/v3/types"
	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/feihua/zero-admin/pkg/chainclient"
)

// 接口实现编译期断言。
var _ chainclient.ChainClient = (*fiscoRealClient)(nil)
var _ chainclient.ChainClient = (disabledClient{})
var _ chainclient.ChainClient = (*invalidConfigClient)(nil)

// disabledClient FISCO 客户端总开关关闭时的占位实现。
type disabledClient struct{}

func (disabledClient) MintToken(context.Context, *chainclient.MintTokenRequest) (*chainclient.MintTokenResponse, error) {
	return nil, errors.New("FISCO BCOS 客户端未配置")
}

func (disabledClient) QueryMintToken(context.Context, *chainclient.QueryMintTokenRequest) (*chainclient.MintTokenResponse, error) {
	return nil, errors.New("FISCO BCOS 客户端未配置")
}

func (disabledClient) ChainType() string { return ChainTypeFisco3x }

// invalidConfigClient Enabled=true 但 Config 校验失败时的占位实现。
// 启动不阻塞，所有调用 fail fast 并附带配置错误原因。
type invalidConfigClient struct {
	cause error
}

func (c *invalidConfigClient) MintToken(context.Context, *chainclient.MintTokenRequest) (*chainclient.MintTokenResponse, error) {
	return nil, &FiscoError{Code: ErrCodeConfigInvalid, Reason: "FISCO 配置非法", Err: c.cause}
}

func (c *invalidConfigClient) QueryMintToken(context.Context, *chainclient.QueryMintTokenRequest) (*chainclient.MintTokenResponse, error) {
	return nil, &FiscoError{Code: ErrCodeConfigInvalid, Reason: "FISCO 配置非法", Err: c.cause}
}

func (c *invalidConfigClient) ChainType() string { return ChainTypeFisco3x }

// fiscoRealClient 真实链客户端。SDK 连接 lazy 创建。
type fiscoRealClient struct {
	cfg          Config
	parsedABI    bcosabi.ABI
	abiJSON      string
	contractAddr common.Address
	privateKey   []byte
	fromAddr     common.Address
	cardMintedID common.Hash // keccak256("CardMinted(address,uint256,string,string)")

	mu        sync.RWMutex
	sdkClient *bcosclient.Client
}

// NewClient 工厂函数：根据 cfg 选择 disabled / invalidConfig / real 客户端。
//
// 行为：
//   - cfg.Enabled=false → disabledClient（错误文案 "FISCO BCOS 客户端未配置"）
//   - cfg.Enabled=true 且 Validate 失败 → invalidConfigClient（启动不崩，调用 fail fast）
//   - cfg.Enabled=true 且 Validate 通过 → fiscoRealClient（lazy SDK 连接）
//
// 不返回 error，保持与旧版相同 signature，方便 ServiceContext 调用方零侵入接入。
func NewClient(cfg Config) chainclient.ChainClient {
	if !cfg.Enabled {
		return disabledClient{}
	}
	if err := cfg.Validate(); err != nil {
		return &invalidConfigClient{cause: err}
	}
	abiJSON, err := cfg.loadContractABI()
	if err != nil {
		return &invalidConfigClient{cause: err}
	}
	parsedABI, err := bcosabi.JSON(strings.NewReader(abiJSON))
	if err != nil {
		return &invalidConfigClient{cause: fmt.Errorf("解析合约 ABI 失败: %w", err)}
	}
	if cfg.IsSMCrypto {
		parsedABI.SetSMCrypto()
	}
	pkBytes, err := cfg.loadPrivateKey()
	if err != nil {
		return &invalidConfigClient{cause: err}
	}
	priv, err := crypto.ToECDSA(pkBytes)
	if err != nil {
		return &invalidConfigClient{cause: fmt.Errorf("私钥派生失败: %w", err)}
	}
	pub, ok := priv.Public().(*ecdsa.PublicKey)
	if !ok {
		return &invalidConfigClient{cause: errors.New("私钥公钥类型异常")}
	}
	fromAddr := crypto.PubkeyToAddress(*pub)

	contractAddr := common.HexToAddress(strings.TrimSpace(cfg.ContractAddr))

	// 预计算 CardMinted event id（topic[0]）
	cardMintedSig := []byte("CardMinted(address,uint256,string,string)")
	cardMintedID := crypto.Keccak256Hash(cardMintedSig)

	return &fiscoRealClient{
		cfg:          cfg,
		parsedABI:    parsedABI,
		abiJSON:      abiJSON,
		contractAddr: contractAddr,
		privateKey:   pkBytes,
		fromAddr:     fromAddr,
		cardMintedID: cardMintedID,
	}
}

// ChainType Story 10.11 明确：使用 "fisco_bcos_3x" 区分真假链。
func (c *fiscoRealClient) ChainType() string { return ChainTypeFisco3x }

// getOrConnect lazy 连接 FISCO 节点。每个 fiscoRealClient 持久维护一个 sdkClient。
func (c *fiscoRealClient) getOrConnect(ctx context.Context) (*bcosclient.Client, error) {
	c.mu.RLock()
	cli := c.sdkClient
	c.mu.RUnlock()
	if cli != nil {
		return cli, nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.sdkClient != nil {
		return c.sdkClient, nil
	}
	sdkCfg := &bcosclient.Config{
		IsSMCrypto:  c.cfg.IsSMCrypto,
		GroupID:     c.cfg.resolveGroupID(),
		Host:        strings.TrimSpace(c.cfg.Host),
		Port:        c.cfg.Port,
		DisableSsl:  c.cfg.DisableSsl,
		PrivateKey:  c.privateKey,
		TLSCaFile:   strings.TrimSpace(c.cfg.CaCertPath),
		TLSCertFile: strings.TrimSpace(c.cfg.SdkCertPath),
		TLSKeyFile:  strings.TrimSpace(c.cfg.SdkKeyPath),
	}
	dialCtx, cancel := context.WithTimeout(ctx, c.cfg.timeout())
	defer cancel()
	cli, err := bcosclient.DialContext(dialCtx, sdkCfg)
	if err != nil {
		return nil, wrapErr(err)
	}
	c.sdkClient = cli
	return cli, nil
}

// MintToken 真实上链铸造 NFT 并解析 tokenId。
func (c *fiscoRealClient) MintToken(ctx context.Context, req *chainclient.MintTokenRequest) (*chainclient.MintTokenResponse, error) {
	if req == nil {
		return nil, errors.New("mint request 不能为空")
	}
	idempotencyKey := strings.TrimSpace(req.IdempotencyKey)
	if idempotencyKey == "" {
		return nil, errors.New("幂等键不能为空")
	}

	cli, err := c.getOrConnect(ctx)
	if err != nil {
		return nil, err
	}

	metadataURI, err := buildMetadataURI(req)
	if err != nil {
		return nil, &FiscoError{Code: ErrCodeABIDecodeFailed, Reason: "metadata 序列化失败", Err: err}
	}

	// ABI Pack mint(to, idempotencyKey, metadataURI)
	calldata, err := c.parsedABI.Pack("mint", c.fromAddr, idempotencyKey, metadataURI)
	if err != nil {
		return nil, &FiscoError{Code: ErrCodeABIDecodeFailed, Reason: "mint calldata 编码失败", Err: err}
	}

	// blockLimit = currentBlock + 500（FISCO 3.x 推荐窗口）
	currentBlock, err := cli.GetBlockNumber(ctx)
	if err != nil {
		return nil, wrapErr(fmt.Errorf("GetBlockNumber: %w", err))
	}
	blockLimit := currentBlock + 500

	receipt, txHash, err := c.sendTx(ctx, cli, &c.contractAddr, calldata, blockLimit)
	if err != nil {
		return nil, err
	}

	if receipt.Status != 0 {
		// receipt.Output 通常包含 revert reason
		revertReason := decodeRevertReason(receipt.Output)
		errMsg := fmt.Sprintf("tx status=%d", receipt.Status)
		if revertReason != "" {
			errMsg = "execution reverted: " + revertReason
		}
		return nil, wrapErr(errors.New(errMsg))
	}

	tokenIDDec, err := c.parseTokenIDFromLogs(receipt.Logs)
	if err != nil {
		return nil, &FiscoError{Code: ErrCodeABIDecodeFailed, Reason: "无法从 Logs 中解析 tokenId", Err: err}
	}

	return c.buildMintResponse(receipt, txHash, tokenIDDec, idempotencyKey, metadataURI), nil
}

// QueryMintToken 通过合约 getTokenByKey(idempotencyKey) 查询 tokenId。
// tokenId == 0 → 返回 chainclient.ErrReceiptNotFound（上游用此信号决定是否走 mint 主流程）。
func (c *fiscoRealClient) QueryMintToken(ctx context.Context, req *chainclient.QueryMintTokenRequest) (*chainclient.MintTokenResponse, error) {
	if req == nil {
		return nil, errors.New("query request 不能为空")
	}
	idempotencyKey := strings.TrimSpace(req.IdempotencyKey)
	if idempotencyKey == "" {
		return nil, errors.New("幂等键不能为空")
	}

	cli, err := c.getOrConnect(ctx)
	if err != nil {
		return nil, err
	}

	calldata, err := c.parsedABI.Pack("getTokenByKey", idempotencyKey)
	if err != nil {
		return nil, &FiscoError{Code: ErrCodeABIDecodeFailed, Reason: "getTokenByKey calldata 编码失败", Err: err}
	}

	callMsg := ethereum.CallMsg{
		To:   &c.contractAddr,
		Data: calldata,
	}
	callCtx, cancel := context.WithTimeout(ctx, c.cfg.timeout())
	defer cancel()
	output, err := cli.CallContract(callCtx, callMsg)
	if err != nil {
		return nil, wrapErr(err)
	}

	var tokenID *big.Int
	if err = c.parsedABI.Unpack(&tokenID, "getTokenByKey", output); err != nil {
		return nil, &FiscoError{Code: ErrCodeABIDecodeFailed, Reason: "getTokenByKey 返回值解码失败", Err: err}
	}
	if tokenID == nil || tokenID.Sign() == 0 {
		return nil, chainclient.ErrReceiptNotFound
	}

	// 当前 Story 不维护回执缓存，仅返回 tokenId + 占位字段。
	// 下个 Story 加 sms_chain_receipt_cache 表后再补 tx hash + 完整回执。
	return &chainclient.MintTokenResponse{
		TokenID:        tokenID.String(),
		ChainTxID:      "", // view 调用无 tx hash
		ChainStatus:    "success",
		ReceiptSummary: fmt.Sprintf("getTokenByKey hit, tokenId=%s", tokenID.String()),
		ReceiptJSON:    fmt.Sprintf(`{"source":"view-call","method":"getTokenByKey","tokenId":"%s","idempotencyKey":%q}`, tokenID.String(), idempotencyKey),
		ConfirmedAt:    time.Now(),
	}, nil
}

// sendTx 执行 mint 交易：构造 → 签名 → 发送 → 同步等回执。
// 返回 (*receipt, txHashHex, error)。
func (c *fiscoRealClient) sendTx(ctx context.Context, cli *bcosclient.Client, to *common.Address, input []byte, blockLimit int64) (*bcostypes.Receipt, string, error) {
	txData, txHash, err := cli.CreateEncodedTransactionDataV1(to, input, blockLimit, c.abiJSON)
	if err != nil {
		return nil, "", wrapErr(fmt.Errorf("CreateEncodedTransactionDataV1: %w", err))
	}
	signature, err := cli.CreateEncodedSignature(txHash)
	if err != nil {
		return nil, "", wrapErr(fmt.Errorf("CreateEncodedSignature: %w", err))
	}
	tx, err := cli.CreateEncodedTransaction(txData, txHash, signature, 0, "")
	if err != nil {
		return nil, "", wrapErr(fmt.Errorf("CreateEncodedTransaction: %w", err))
	}

	sendCtx, cancel := context.WithTimeout(ctx, c.cfg.timeout())
	defer cancel()
	receipt, err := cli.SendEncodedTransaction(sendCtx, tx, true)
	if err != nil {
		return nil, "", wrapErr(fmt.Errorf("SendEncodedTransaction: %w", err))
	}
	if receipt == nil {
		return nil, "", wrapErr(errors.New("receipt is nil"))
	}
	txHashHex := "0x" + hex.EncodeToString(txHash)
	return receipt, txHashHex, nil
}

// parseTokenIDFromLogs 从 receipt.Logs 中匹配 CardMinted event 并提取 tokenId（十进制）。
//
// CardMinted(address indexed to, uint256 indexed tokenId, string idempotencyKey, string metadataURI)
//
// topic[0] = keccak256(signature)
// topic[1] = to (32 字节左 padded)
// topic[2] = tokenId (32 字节大端 uint256)
//
// 注：FISCO go-sdk v3 的 NewLog 中 Topics/Data 都是 hex 字符串而非 bytes，需要先 decode。
func (c *fiscoRealClient) parseTokenIDFromLogs(logs []*bcostypes.NewLog) (string, error) {
	for _, lg := range logs {
		if lg == nil || len(lg.Topics) < 3 {
			continue
		}
		topic0Bytes, err := hex.DecodeString(strings.TrimPrefix(strings.TrimPrefix(lg.Topics[0], "0x"), "0X"))
		if err != nil || !equalBytes(topic0Bytes, c.cardMintedID.Bytes()) {
			continue
		}
		topic2Bytes, err := hex.DecodeString(strings.TrimPrefix(strings.TrimPrefix(lg.Topics[2], "0x"), "0X"))
		if err != nil {
			return "", fmt.Errorf("topic[2] hex 解码失败: %w", err)
		}
		tokenID := new(big.Int).SetBytes(topic2Bytes)
		return tokenID.String(), nil
	}
	return "", errors.New("CardMinted event 未在 receipt.Logs 中找到")
}

// buildMintResponse 构造 chainclient.MintTokenResponse。
// Story 5.4.7 要求 ReceiptSummary 含 block + gasUsed，ReceiptJSON 是完整 receipt。
func (c *fiscoRealClient) buildMintResponse(receipt *bcostypes.Receipt, txHash, tokenIDDec, idempotencyKey, metadataURI string) *chainclient.MintTokenResponse {
	confirmedAt := time.Now()

	receiptJSON := encodeReceiptJSON(receipt, txHash, tokenIDDec, idempotencyKey, metadataURI)
	summary := fmt.Sprintf("block=%d gasUsed=%s tokenId=%s", receipt.BlockNumber, strings.TrimSpace(receipt.GasUsed), tokenIDDec)

	return &chainclient.MintTokenResponse{
		TokenID:        tokenIDDec,
		ChainTxID:      txHash,
		ChainStatus:    "success",
		ReceiptSummary: summary,
		ReceiptJSON:    receiptJSON,
		ConfirmedAt:    confirmedAt,
	}
}

// buildMetadataURI 把请求关键字段序列化为 metadata JSON 串
// （本 Story 暂用 inline JSON，下一 Story 改成 IPFS / 后端 API URL）。
func buildMetadataURI(req *chainclient.MintTokenRequest) (string, error) {
	payload := map[string]interface{}{
		"idempotencyKey":  strings.TrimSpace(req.IdempotencyKey),
		"taskId":          req.TaskID,
		"assetInstanceId": req.AssetInstanceID,
		"activityId":      req.ActivityID,
		"templateId":      req.TemplateID,
		"memberId":        req.MemberID,
		"requestId":       strings.TrimSpace(req.RequestID),
		"traceId":         strings.TrimSpace(req.TraceID),
		"assetNo":         strings.TrimSpace(req.AssetNo),
		"scopeType":       strings.TrimSpace(req.ScopeType),
		"platformId":      req.PlatformID,
		"tenantId":        req.TenantID,
		"merchantId":      req.MerchantID,
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return "data:application/json;charset=utf-8," + string(b), nil
}

// encodeReceiptJSON 把 SDK Receipt 序列化为完整 JSON（用于 sms_card_mint_task.last_receipt_json）。
// 不直接 json.Marshal(receipt) 因为 receipt.Logs 中含 BigInt 等类型；这里挑关键字段。
func encodeReceiptJSON(receipt *bcostypes.Receipt, txHash, tokenIDDec, idempotencyKey, metadataURI string) string {
	logs := logs0(receipt.Logs)
	payload := map[string]interface{}{
		"chainType":      ChainTypeFisco3x,
		"txHash":         txHash,
		"tokenId":        tokenIDDec,
		"idempotencyKey": idempotencyKey,
		"metadataUri":    metadataURI,
		"blockNumber":    receipt.BlockNumber,
		"contractAddr":   receipt.ContractAddress,
		"from":           receipt.From,
		"to":             receipt.To,
		"gasUsed":        receipt.GasUsed,
		"status":         receipt.Status,
		"output":         receipt.Output,
		"logs":           logs,
	}
	b, err := json.Marshal(payload)
	if err != nil {
		// 兜底：拼字符串保 ChainTxID 不丢
		return fmt.Sprintf(`{"chainType":%q,"txHash":%q,"tokenId":%q,"marshalErr":%q}`,
			ChainTypeFisco3x, txHash, tokenIDDec, err.Error())
	}
	return string(b)
}

// logs0 把 SDK NewLog 列表转成可 JSON 序列化的 map 切片。
func logs0(in []*bcostypes.NewLog) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(in))
	for _, lg := range in {
		if lg == nil {
			continue
		}
		out = append(out, map[string]interface{}{
			"address":     lg.Address,
			"topics":      lg.Topics,
			"data":        lg.Data,
			"blockNumber": lg.BlockNumber,
		})
	}
	return out
}

// decodeRevertReason 从 EVM revert output 中解析 reason 字符串。
// EVM revert 编码为 Error(string)：
//
//	bytes[0:4]  = keccak256("Error(string)")[:4]
//	bytes[4:36] = ABI offset (固定 0x20)
//	bytes[36:68]= string length (大端 uint256)
//	bytes[68:]  = string bytes (UTF-8)
func decodeRevertReason(output string) string {
	output = strings.TrimPrefix(strings.TrimPrefix(output, "0x"), "0X")
	if output == "" {
		return ""
	}
	raw, err := hex.DecodeString(output)
	if err != nil {
		return output
	}
	if len(raw) < 68 {
		return ""
	}
	// 校验 selector
	expectedSelector := crypto.Keccak256Hash([]byte("Error(string)")).Bytes()[:4]
	if !equalBytes(raw[:4], expectedSelector) {
		return ""
	}
	strLen := new(big.Int).SetBytes(raw[36:68]).Int64()
	if strLen <= 0 || int(strLen)+68 > len(raw) {
		return ""
	}
	return string(raw[68 : 68+strLen])
}

func equalBytes(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
