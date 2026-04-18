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
)

type Client interface {
	MintToken(ctx context.Context, req *MintTokenRequest) (*MintTokenResponse, error)
}

type disabledClient struct{}

func (disabledClient) MintToken(context.Context, *MintTokenRequest) (*MintTokenResponse, error) {
	return nil, errors.New("蚂蚁链能力未启用")
}

type httpClient struct {
	cfg        Config
	httpClient *http.Client
}

type mintTokenHTTPPayload struct {
	AppID          string `json:"appId"`
	AccessKey      string `json:"accessKey"`
	Secret         string `json:"secret"`
	IdempotencyKey string `json:"idempotencyKey"`
	TaskID         int64  `json:"taskId"`
	AssetInstanceID int64 `json:"assetInstanceId"`
	ActivityID     int64  `json:"activityId"`
	TemplateID     int64  `json:"templateId"`
	MemberID       int64  `json:"memberId"`
	RequestID      string `json:"requestId"`
	TraceID        string `json:"traceId"`
	AssetNo        string `json:"assetNo"`
	ScopeType      string `json:"scopeType"`
	PlatformID     int64  `json:"platformId"`
	TenantID       int64  `json:"tenantId"`
	MerchantID     int64  `json:"merchantId"`
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

func NewClient(cfg Config) Client {
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

func (c *httpClient) MintToken(ctx context.Context, req *MintTokenRequest) (*MintTokenResponse, error) {
	if req == nil {
		return nil, errors.New("mint request 不能为空")
	}

	payload := mintTokenHTTPPayload{
		AppID:          strings.TrimSpace(c.cfg.AppID),
		AccessKey:      strings.TrimSpace(c.cfg.AccessKey),
		Secret:         strings.TrimSpace(c.cfg.Secret),
		IdempotencyKey: strings.TrimSpace(req.IdempotencyKey),
		TaskID:         req.TaskID,
		AssetInstanceID: req.AssetInstanceID,
		ActivityID:     req.ActivityID,
		TemplateID:     req.TemplateID,
		MemberID:       req.MemberID,
		RequestID:      strings.TrimSpace(req.RequestID),
		TraceID:        strings.TrimSpace(req.TraceID),
		AssetNo:        strings.TrimSpace(req.AssetNo),
		ScopeType:      strings.TrimSpace(req.ScopeType),
		PlatformID:     req.PlatformID,
		TenantID:       req.TenantID,
		MerchantID:     req.MerchantID,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimSpace(c.cfg.Endpoint), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= http.StatusBadRequest {
		return nil, fmt.Errorf("蚂蚁链请求失败: http %d %s", resp.StatusCode, strings.TrimSpace(string(raw)))
	}

	var result mintTokenHTTPResponse
	if err = json.Unmarshal(raw, &result); err != nil {
		return nil, err
	}
	if strings.TrimSpace(result.Code) != "" && result.Code != "0" && !strings.EqualFold(result.Code, "success") {
		return nil, fmt.Errorf("蚂蚁链发放失败: %s", firstNonEmpty(result.Message, result.Code))
	}

	confirmedAt := time.Now()
	if parsed, parseErr := time.Parse(time.RFC3339, strings.TrimSpace(result.ConfirmedAt)); parseErr == nil {
		confirmedAt = parsed
	}

	return &MintTokenResponse{
		TokenID:        strings.TrimSpace(result.TokenID),
		ChainTxID:      strings.TrimSpace(result.ChainTxID),
		ChainStatus:    firstNonEmpty(strings.TrimSpace(result.ChainStatus), "success"),
		ReceiptSummary: strings.TrimSpace(result.ReceiptSummary),
		ReceiptJSON:    strings.TrimSpace(result.ReceiptJSON),
		ConfirmedAt:    confirmedAt,
	}, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

