package antchain

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/feihua/zero-admin/pkg/chainclient"
)

type Client = chainclient.ChainClient

var _ chainclient.ChainClient = (*httpClient)(nil)
var _ chainclient.ChainClient = (disabledClient{})

type disabledClient struct{}

func (disabledClient) MintToken(context.Context, *chainclient.MintTokenRequest) (*chainclient.MintTokenResponse, error) {
	return nil, errors.New("蚂蚁链能力未启用")
}

func (disabledClient) QueryMintToken(context.Context, *chainclient.QueryMintTokenRequest) (*chainclient.MintTokenResponse, error) {
	return nil, errors.New("蚂蚁链能力未启用")
}

func (disabledClient) ChainType() string { return "antchain" }

type httpClient struct {
	cfg        Config
	httpClient *http.Client
}

type mintTokenHTTPPayload struct {
	AppID           string `json:"appId"`
	AccessKey       string `json:"accessKey"`
	Secret          string `json:"secret"`
	IdempotencyKey  string `json:"idempotencyKey"`
	TaskID          int64  `json:"taskId"`
	AssetInstanceID int64  `json:"assetInstanceId"`
	ActivityID      int64  `json:"activityId"`
	TemplateID      int64  `json:"templateId"`
	MemberID        int64  `json:"memberId"`
	RequestID       string `json:"requestId"`
	TraceID         string `json:"traceId"`
	AssetNo         string `json:"assetNo"`
	ScopeType       string `json:"scopeType"`
	PlatformID      int64  `json:"platformId"`
	TenantID        int64  `json:"tenantId"`
	MerchantID      int64  `json:"merchantId"`
}

type queryMintTokenHTTPPayload struct {
	AppID           string `json:"appId"`
	AccessKey       string `json:"accessKey"`
	Secret          string `json:"secret"`
	IdempotencyKey  string `json:"idempotencyKey"`
	TaskID          int64  `json:"taskId"`
	AssetInstanceID int64  `json:"assetInstanceId"`
	RequestID       string `json:"requestId"`
	TraceID         string `json:"traceId"`
}

type mintTokenHTTPResponse struct {
	Code           string `json:"code"`
	Message        string `json:"message"`
	TokenID        string `json:"tokenId"`
	ChainTxID      string `json:"chainTxId"`
	ChainStatus    string `json:"chainStatus"`
	ReceiptSummary string `json:"receiptSummary"`
	ReceiptJSON    string `json:"receiptJson"`
	ConfirmedAt    string `json:"confirmedAt"`
}

type queryMintTokenHTTPResponse struct {
	Code           string `json:"code"`
	Message        string `json:"message"`
	Found          bool   `json:"found"`
	TokenID        string `json:"tokenId"`
	ChainTxID      string `json:"chainTxId"`
	ChainStatus    string `json:"chainStatus"`
	ReceiptSummary string `json:"receiptSummary"`
	ReceiptJSON    string `json:"receiptJson"`
	ConfirmedAt    string `json:"confirmedAt"`
}

func NewClient(cfg Config) chainclient.ChainClient {
	if !cfg.Enabled || strings.TrimSpace(cfg.Endpoint) == "" {
		return disabledClient{}
	}
	timeout := time.Duration(cfg.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	return &httpClient{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (c *httpClient) ChainType() string { return "antchain" }

func (c *httpClient) MintToken(ctx context.Context, req *chainclient.MintTokenRequest) (*chainclient.MintTokenResponse, error) {
	if req == nil {
		return nil, errors.New("mint request 不能为空")
	}

	payload := mintTokenHTTPPayload{
		AppID:           strings.TrimSpace(c.cfg.AppID),
		AccessKey:       strings.TrimSpace(c.cfg.AccessKey),
		Secret:          strings.TrimSpace(c.cfg.Secret),
		IdempotencyKey:  strings.TrimSpace(req.IdempotencyKey),
		TaskID:          req.TaskID,
		AssetInstanceID: req.AssetInstanceID,
		ActivityID:      req.ActivityID,
		TemplateID:      req.TemplateID,
		MemberID:        req.MemberID,
		RequestID:       strings.TrimSpace(req.RequestID),
		TraceID:         strings.TrimSpace(req.TraceID),
		AssetNo:         strings.TrimSpace(req.AssetNo),
		ScopeType:       strings.TrimSpace(req.ScopeType),
		PlatformID:      req.PlatformID,
		TenantID:        req.TenantID,
		MerchantID:      req.MerchantID,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	raw, statusCode, err := c.doJSONRequest(ctx, strings.TrimSpace(c.cfg.Endpoint), body)
	if err != nil {
		return nil, err
	}
	if statusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("蚂蚁链请求失败: http %d %s", statusCode, strings.TrimSpace(string(raw)))
	}
	return parseMintTokenResponse(raw)
}

func (c *httpClient) QueryMintToken(ctx context.Context, req *chainclient.QueryMintTokenRequest) (*chainclient.MintTokenResponse, error) {
	if req == nil {
		return nil, errors.New("query request 不能为空")
	}

	payload := queryMintTokenHTTPPayload{
		AppID:           strings.TrimSpace(c.cfg.AppID),
		AccessKey:       strings.TrimSpace(c.cfg.AccessKey),
		Secret:          strings.TrimSpace(c.cfg.Secret),
		IdempotencyKey:  strings.TrimSpace(req.IdempotencyKey),
		TaskID:          req.TaskID,
		AssetInstanceID: req.AssetInstanceID,
		RequestID:       strings.TrimSpace(req.RequestID),
		TraceID:         strings.TrimSpace(req.TraceID),
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	raw, statusCode, err := c.doJSONRequest(ctx, c.receiptEndpoint(), body)
	if err != nil {
		return nil, err
	}
	if statusCode == http.StatusNotFound {
		return nil, ErrReceiptNotFound
	}
	if statusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("蚂蚁链回执查询失败: http %d %s", statusCode, strings.TrimSpace(string(raw)))
	}

	var result queryMintTokenHTTPResponse
	if err = json.Unmarshal(raw, &result); err != nil {
		return nil, err
	}
	if codeIndicatesNotFound(result.Code) {
		return nil, ErrReceiptNotFound
	}
	if strings.TrimSpace(result.Code) != "" && result.Code != "0" && !strings.EqualFold(result.Code, "success") {
		return nil, fmt.Errorf("蚂蚁链回执查询失败: %s", firstNonEmpty(result.Message, result.Code))
	}
	if !result.Found && strings.TrimSpace(result.TokenID) == "" {
		return nil, ErrReceiptNotFound
	}
	return buildMintTokenResponse(
		result.TokenID,
		result.ChainTxID,
		result.ChainStatus,
		result.ReceiptSummary,
		result.ReceiptJSON,
		result.ConfirmedAt,
	), nil
}

func (c *httpClient) doJSONRequest(ctx context.Context, endpoint string, body []byte) ([]byte, int, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimSpace(endpoint), bytes.NewReader(body))
	if err != nil {
		return nil, 0, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, 0, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, err
	}
	return raw, resp.StatusCode, nil
}

func (c *httpClient) receiptEndpoint() string {
	return firstNonEmpty(strings.TrimSpace(c.cfg.ReceiptEndpoint), strings.TrimSpace(c.cfg.Endpoint))
}

func parseMintTokenResponse(raw []byte) (*chainclient.MintTokenResponse, error) {
	var result mintTokenHTTPResponse
	if err := json.Unmarshal(raw, &result); err != nil {
		return nil, err
	}
	if strings.TrimSpace(result.Code) != "" && result.Code != "0" && !strings.EqualFold(result.Code, "success") {
		return nil, fmt.Errorf("蚂蚁链发放失败: %s", firstNonEmpty(result.Message, result.Code))
	}
	return buildMintTokenResponse(
		result.TokenID,
		result.ChainTxID,
		result.ChainStatus,
		result.ReceiptSummary,
		result.ReceiptJSON,
		result.ConfirmedAt,
	), nil
}

func buildMintTokenResponse(tokenID, chainTxID, chainStatus, receiptSummary, receiptJSON, confirmedAtRaw string) *chainclient.MintTokenResponse {
	confirmedAt := time.Now()
	if parsed, parseErr := time.Parse(time.RFC3339, strings.TrimSpace(confirmedAtRaw)); parseErr == nil {
		confirmedAt = parsed
	}
	return &chainclient.MintTokenResponse{
		TokenID:        strings.TrimSpace(tokenID),
		ChainTxID:      strings.TrimSpace(chainTxID),
		ChainStatus:    firstNonEmpty(strings.TrimSpace(chainStatus), "success"),
		ReceiptSummary: strings.TrimSpace(receiptSummary),
		ReceiptJSON:    strings.TrimSpace(receiptJSON),
		ConfirmedAt:    confirmedAt,
	}
}

func codeIndicatesNotFound(code string) bool {
	switch strings.ToLower(strings.TrimSpace(code)) {
	case "404", "not_found", "receipt_not_found":
		return true
	default:
		return false
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
