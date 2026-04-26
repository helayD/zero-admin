package fisco

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/feihua/zero-admin/pkg/chainclient"
)

func TestNewClientDisabled(t *testing.T) {
	client := NewClient(Config{Enabled: false})

	_, err := client.MintToken(context.Background(), &chainclient.MintTokenRequest{IdempotencyKey: "k1"})
	if err == nil || !strings.Contains(err.Error(), "免费链能力未启用") {
		t.Fatalf("expected disabled error for MintToken, got %v", err)
	}

	_, err = client.QueryMintToken(context.Background(), &chainclient.QueryMintTokenRequest{IdempotencyKey: "k1"})
	if err == nil || !strings.Contains(err.Error(), "免费链能力未启用") {
		t.Fatalf("expected disabled error for QueryMintToken, got %v", err)
	}
}

func TestNewClientEnabled(t *testing.T) {
	client := NewClient(Config{Enabled: true, NodeAddr: "127.0.0.1:20200"})
	if client.ChainType() != "free_chain" {
		t.Fatalf("expected ChainType free_chain, got %s", client.ChainType())
	}
}

func TestEnabledClientMintsAndQueriesStableReceipt(t *testing.T) {
	client := NewClient(Config{Enabled: true})
	req := &chainclient.MintTokenRequest{
		IdempotencyKey:  "card-mint:970001",
		TaskID:          980001,
		AssetInstanceID: 970001,
		ActivityID:      930001,
		TemplateID:      920001,
		MemberID:        3001,
		RequestID:       "req-free-1",
		TraceID:         "trace-free-1",
		AssetNo:         "ASSET-FREE-001",
		ScopeType:       "merchant",
		PlatformID:      1,
		TenantID:        10,
		MerchantID:      88,
	}

	first, err := client.MintToken(context.Background(), req)
	if err != nil {
		t.Fatalf("MintToken returned error: %v", err)
	}
	second, err := client.MintToken(context.Background(), req)
	if err != nil {
		t.Fatalf("second MintToken returned error: %v", err)
	}
	if first.TokenID == "" || first.ChainTxID == "" || first.ReceiptSummary == "" || first.ReceiptJSON == "" {
		t.Fatalf("expected stable receipt fields, got %+v", first)
	}
	if first.TokenID != second.TokenID || first.ChainTxID != second.ChainTxID || first.ReceiptJSON != second.ReceiptJSON {
		t.Fatalf("expected idempotent response, first=%+v second=%+v", first, second)
	}

	queried, err := client.QueryMintToken(context.Background(), &chainclient.QueryMintTokenRequest{
		IdempotencyKey:  req.IdempotencyKey,
		TaskID:          req.TaskID,
		AssetInstanceID: req.AssetInstanceID,
		RequestID:       req.RequestID,
		TraceID:         req.TraceID,
	})
	if err != nil {
		t.Fatalf("QueryMintToken returned error: %v", err)
	}
	if queried.TokenID != first.TokenID || queried.ChainTxID != first.ChainTxID {
		t.Fatalf("expected query to return minted receipt, got %+v want %+v", queried, first)
	}

	var payload map[string]interface{}
	if err = json.Unmarshal([]byte(first.ReceiptJSON), &payload); err != nil {
		t.Fatalf("receipt json is invalid: %v", err)
	}
	if payload["chainType"] != "free_chain" || payload["assetNo"] != req.AssetNo {
		t.Fatalf("unexpected receipt payload: %+v", payload)
	}
}

func TestFiscoClientImplementsChainClient(t *testing.T) {
	var _ chainclient.ChainClient = (*fiscoClient)(nil)
	var _ chainclient.ChainClient = (disabledClient{})
}
