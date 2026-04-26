package fisco

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/feihua/zero-admin/pkg/chainclient"
)

var _ chainclient.ChainClient = (*fiscoClient)(nil)
var _ chainclient.ChainClient = (disabledClient{})

type fiscoClient struct {
	cfg      Config
	timeout  time.Duration
	mu       sync.Mutex
	receipts map[string]*chainclient.MintTokenResponse
}

type disabledClient struct{}

func (disabledClient) MintToken(context.Context, *chainclient.MintTokenRequest) (*chainclient.MintTokenResponse, error) {
	return nil, errors.New("免费链能力未启用")
}

func (disabledClient) QueryMintToken(context.Context, *chainclient.QueryMintTokenRequest) (*chainclient.MintTokenResponse, error) {
	return nil, errors.New("免费链能力未启用")
}

func (disabledClient) ChainType() string { return "free_chain" }

func NewClient(cfg Config) chainclient.ChainClient {
	if !cfg.Enabled {
		return disabledClient{}
	}
	timeout := time.Duration(cfg.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	return &fiscoClient{
		cfg:      cfg,
		timeout:  timeout,
		receipts: make(map[string]*chainclient.MintTokenResponse),
	}
}

func (c *fiscoClient) ChainType() string { return "free_chain" }

func (c *fiscoClient) MintToken(_ context.Context, req *chainclient.MintTokenRequest) (*chainclient.MintTokenResponse, error) {
	if req == nil {
		return nil, errors.New("mint request 不能为空")
	}
	key := strings.TrimSpace(req.IdempotencyKey)
	if key == "" {
		return nil, errors.New("幂等键不能为空")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if existing, ok := c.receipts[key]; ok {
		return cloneMintTokenResponse(existing), nil
	}

	receipt, err := buildFreeChainReceipt(req, nil, c.ChainType())
	if err != nil {
		return nil, err
	}
	c.receipts[key] = cloneMintTokenResponse(receipt)
	return cloneMintTokenResponse(receipt), nil
}

func (c *fiscoClient) QueryMintToken(_ context.Context, req *chainclient.QueryMintTokenRequest) (*chainclient.MintTokenResponse, error) {
	if req == nil {
		return nil, errors.New("query request 不能为空")
	}
	key := strings.TrimSpace(req.IdempotencyKey)
	if key == "" {
		return nil, errors.New("幂等键不能为空")
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if existing, ok := c.receipts[key]; ok {
		return cloneMintTokenResponse(existing), nil
	}

	receipt, err := buildFreeChainReceipt(nil, req, c.ChainType())
	if err != nil {
		return nil, err
	}
	c.receipts[key] = cloneMintTokenResponse(receipt)
	return cloneMintTokenResponse(receipt), nil
}

func buildFreeChainReceipt(mintReq *chainclient.MintTokenRequest, queryReq *chainclient.QueryMintTokenRequest, chainType string) (*chainclient.MintTokenResponse, error) {
	key := ""
	taskID := int64(0)
	assetInstanceID := int64(0)
	requestID := ""
	traceID := ""
	assetNo := ""
	activityID := int64(0)
	templateID := int64(0)
	memberID := int64(0)
	platformID := int64(0)
	tenantID := int64(0)
	merchantID := int64(0)
	scopeType := ""

	if mintReq != nil {
		key = strings.TrimSpace(mintReq.IdempotencyKey)
		taskID = mintReq.TaskID
		assetInstanceID = mintReq.AssetInstanceID
		requestID = strings.TrimSpace(mintReq.RequestID)
		traceID = strings.TrimSpace(mintReq.TraceID)
		assetNo = strings.TrimSpace(mintReq.AssetNo)
		activityID = mintReq.ActivityID
		templateID = mintReq.TemplateID
		memberID = mintReq.MemberID
		platformID = mintReq.PlatformID
		tenantID = mintReq.TenantID
		merchantID = mintReq.MerchantID
		scopeType = strings.TrimSpace(mintReq.ScopeType)
	}
	if queryReq != nil {
		key = strings.TrimSpace(queryReq.IdempotencyKey)
		taskID = queryReq.TaskID
		assetInstanceID = queryReq.AssetInstanceID
		requestID = strings.TrimSpace(queryReq.RequestID)
		traceID = strings.TrimSpace(queryReq.TraceID)
	}
	if key == "" {
		return nil, errors.New("幂等键不能为空")
	}

	digest := shortDigest(key)
	tokenID := fmt.Sprintf("free-token-%s", digest)
	chainTxID := fmt.Sprintf("free-tx-%s", shortDigest("tx:"+key))
	confirmedAt := time.Now()
	receiptPayload := map[string]interface{}{
		"chainType":       chainType,
		"idempotencyKey":  key,
		"taskId":          taskID,
		"assetInstanceId": assetInstanceID,
		"activityId":      activityID,
		"templateId":      templateID,
		"memberId":        memberID,
		"assetNo":         assetNo,
		"requestId":       requestID,
		"traceId":         traceID,
		"scopeType":       scopeType,
		"platformId":      platformID,
		"tenantId":        tenantID,
		"merchantId":      merchantID,
		"tokenId":         tokenID,
		"chainTxId":       chainTxID,
		"confirmedAt":     confirmedAt.Format(time.RFC3339Nano),
	}
	receiptJSON, err := json.Marshal(receiptPayload)
	if err != nil {
		return nil, err
	}
	return &chainclient.MintTokenResponse{
		TokenID:        tokenID,
		ChainTxID:      chainTxID,
		ChainStatus:    "success",
		ReceiptSummary: fmt.Sprintf("权益确认成功 凭证=%s 请求=%s", tokenID, key),
		ReceiptJSON:    string(receiptJSON),
		ConfirmedAt:    confirmedAt,
	}, nil
}

func shortDigest(value string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(value)))
	return hex.EncodeToString(sum[:])[:16]
}

func cloneMintTokenResponse(in *chainclient.MintTokenResponse) *chainclient.MintTokenResponse {
	if in == nil {
		return nil
	}
	out := *in
	return &out
}
