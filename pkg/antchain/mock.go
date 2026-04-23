package antchain

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/feihua/zero-admin/pkg/chainclient"
)

var _ chainclient.ChainClient = (*MockClient)(nil)

type MockClient struct {
	mu               sync.Mutex
	responses        map[string]*chainclient.MintTokenResponse
	forceErrors      map[string]error
	forceQueryErrors map[string]error
}

func NewMockClient() *MockClient {
	return &MockClient{
		responses:        make(map[string]*chainclient.MintTokenResponse),
		forceErrors:      make(map[string]error),
		forceQueryErrors: make(map[string]error),
	}
}

func (m *MockClient) ChainType() string { return "antchain-mock" }

func (m *MockClient) SetError(idempotencyKey string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.forceErrors[strings.TrimSpace(idempotencyKey)] = err
}

func (m *MockClient) SetQueryError(idempotencyKey string, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.forceQueryErrors[strings.TrimSpace(idempotencyKey)] = err
}

func (m *MockClient) MintToken(_ context.Context, req *chainclient.MintTokenRequest) (*chainclient.MintTokenResponse, error) {
	if req == nil {
		return nil, errors.New("mint request 不能为空")
	}

	key := strings.TrimSpace(req.IdempotencyKey)
	if key == "" {
		return nil, errors.New("幂等键不能为空")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if err, ok := m.forceErrors[key]; ok {
		return nil, err
	}
	if existing, ok := m.responses[key]; ok {
		return cloneMintTokenResponse(existing), nil
	}

	receiptPayload := map[string]interface{}{
		"idempotencyKey":  key,
		"taskId":          req.TaskID,
		"assetInstanceId": req.AssetInstanceID,
		"traceId":         req.TraceID,
		"requestId":       req.RequestID,
	}
	receiptJSON, _ := json.Marshal(receiptPayload)
	now := time.Now()
	resp := &chainclient.MintTokenResponse{
		TokenID:        fmt.Sprintf("token-%d", req.AssetInstanceID),
		ChainTxID:      fmt.Sprintf("tx-%d", req.TaskID),
		ChainStatus:    "success",
		ReceiptSummary: fmt.Sprintf("token=%s tx=%s", fmt.Sprintf("token-%d", req.AssetInstanceID), fmt.Sprintf("tx-%d", req.TaskID)),
		ReceiptJSON:    string(receiptJSON),
		ConfirmedAt:    now,
	}
	m.responses[key] = cloneMintTokenResponse(resp)
	return cloneMintTokenResponse(resp), nil
}

func (m *MockClient) QueryMintToken(_ context.Context, req *chainclient.QueryMintTokenRequest) (*chainclient.MintTokenResponse, error) {
	if req == nil {
		return nil, errors.New("query request 不能为空")
	}

	key := strings.TrimSpace(req.IdempotencyKey)
	if key == "" {
		return nil, errors.New("幂等键不能为空")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if err, ok := m.forceQueryErrors[key]; ok {
		return nil, err
	}
	if existing, ok := m.responses[key]; ok {
		return cloneMintTokenResponse(existing), nil
	}
	return nil, ErrReceiptNotFound
}

func cloneMintTokenResponse(in *chainclient.MintTokenResponse) *chainclient.MintTokenResponse {
	if in == nil {
		return nil
	}
	out := *in
	return &out
}
