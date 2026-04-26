package svc

import (
	"context"
	"strings"
	"testing"

	"github.com/feihua/zero-admin/pkg/chainclient"
	"github.com/feihua/zero-admin/rpc/sms/internal/config"
)

func TestBuildChainClientDefaultsToFreeChain(t *testing.T) {
	var c config.Config
	c.Fisco.Enabled = true

	client := buildChainClient(c)
	if client == nil {
		t.Fatal("expected chain client")
	}
	if client.ChainType() != "free_chain" {
		t.Fatalf("expected free_chain client, got %s", client.ChainType())
	}
}

func TestBuildChainClientRejectsInvalidPrimary(t *testing.T) {
	var c config.Config
	c.Blockchain.Primary = "unknown"

	client := buildChainClient(c)
	if client == nil {
		t.Fatal("expected explicit invalid client")
	}
	_, err := client.MintToken(context.Background(), &chainclient.MintTokenRequest{IdempotencyKey: "k1"})
	if err == nil || !strings.Contains(err.Error(), "数字资产通道配置非法") {
		t.Fatalf("expected invalid primary error, got %v", err)
	}
}
