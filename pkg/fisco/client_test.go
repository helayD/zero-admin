package fisco

import (
	"context"
	"strings"
	"testing"

	"github.com/feihua/zero-admin/pkg/chainclient"
)

func TestNewClientDisabled(t *testing.T) {
	client := NewClient(Config{Enabled: false})

	_, err := client.MintToken(context.Background(), &chainclient.MintTokenRequest{IdempotencyKey: "k1"})
	if err == nil || !strings.Contains(err.Error(), "FISCO BCOS 能力未启用") {
		t.Fatalf("expected disabled error for MintToken, got %v", err)
	}

	_, err = client.QueryMintToken(context.Background(), &chainclient.QueryMintTokenRequest{IdempotencyKey: "k1"})
	if err == nil || !strings.Contains(err.Error(), "FISCO BCOS 能力未启用") {
		t.Fatalf("expected disabled error for QueryMintToken, got %v", err)
	}
}

func TestNewClientEnabled(t *testing.T) {
	client := NewClient(Config{Enabled: true, NodeAddr: "127.0.0.1:20200"})
	if client.ChainType() != "fisco" {
		t.Fatalf("expected ChainType fisco, got %s", client.ChainType())
	}
}

func TestFiscoClientImplementsChainClient(t *testing.T) {
	var _ chainclient.ChainClient = (*fiscoClient)(nil)
	var _ chainclient.ChainClient = (disabledClient{})
}
