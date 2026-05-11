package svc

import (
	"context"
	"strings"
	"testing"

	"github.com/feihua/zero-admin/pkg/chainclient"
	"github.com/feihua/zero-admin/rpc/sms/internal/config"
)

// Story 10.11 / Task 5.6 + 8.4：默认 chainType 必须是 fisco_bcos_3x，禁止再出现 free_chain mock。
func TestBuildChainClientDefaultsToFiscoBcos3x(t *testing.T) {
	var c config.Config
	c.Fisco.Enabled = true

	client := buildChainClient(c)
	if client == nil {
		t.Fatal("expected chain client")
	}
	if client.ChainType() != "fisco_bcos_3x" {
		t.Fatalf("expected fisco_bcos_3x client, got %s", client.ChainType())
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
	if err == nil || !strings.Contains(err.Error(), "提货卡通道配置非法") {
		t.Fatalf("expected invalid primary error, got %v", err)
	}
}
