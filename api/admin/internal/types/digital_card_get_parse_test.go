package types

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/zeromicro/go-zero/rest/httpx"
)

func TestQueryDigitalCardChainListReqParsesGETQuery(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/sms/digitalCardChain/queryDigitalCardChainList?scopeType=platform&platformId=1&pageSize=20&current=1", nil)

	var parsed QueryDigitalCardChainListReq
	if err := httpx.Parse(req, &parsed); err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if parsed.ScopeType != "platform" {
		t.Fatalf("expected scopeType=platform, got %q", parsed.ScopeType)
	}
	if parsed.PlatformId != 1 {
		t.Fatalf("expected platformId=1, got %d", parsed.PlatformId)
	}
}

func TestQueryDigitalCardChainDetailReqParsesGETQuery(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/sms/digitalCardChain/queryDigitalCardChainDetail?taskId=980001", nil)

	var parsed QueryDigitalCardChainDetailReq
	if err := httpx.Parse(req, &parsed); err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if parsed.TaskId != 980001 {
		t.Fatalf("expected taskId=980001, got %d", parsed.TaskId)
	}
}

func TestQueryDigitalCardAssetListReqParsesGETQuery(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/sms/digitalCardAsset/queryDigitalCardAssetList?scopeType=platform&platformId=1&pageSize=20&current=1", nil)

	var parsed QueryDigitalCardAssetListReq
	if err := httpx.Parse(req, &parsed); err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if parsed.ScopeType != "platform" {
		t.Fatalf("expected scopeType=platform, got %q", parsed.ScopeType)
	}
	if parsed.PlatformId != 1 {
		t.Fatalf("expected platformId=1, got %d", parsed.PlatformId)
	}
}

func TestQueryDigitalCardAssetDetailReqParsesGETQuery(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/sms/digitalCardAsset/queryDigitalCardAssetDetail?assetInstanceId=970001", nil)

	var parsed QueryDigitalCardAssetDetailReq
	if err := httpx.Parse(req, &parsed); err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if parsed.AssetInstanceId != 970001 {
		t.Fatalf("expected assetInstanceId=970001, got %d", parsed.AssetInstanceId)
	}
}

func TestDigitalCardAssetActionReqParsesJSONBody(t *testing.T) {
	req := httptest.NewRequest("POST", "/api/sms/digitalCardAsset/reviewDigitalCardAssetCompliance", strings.NewReader(`{"assetInstanceId":970001,"reason":"manual review","scopeType":"platform","platformId":1}`))
	req.Header.Set("Content-Type", "application/json")

	var parsed DigitalCardAssetActionReq
	if err := httpx.Parse(req, &parsed); err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if parsed.ScopeType != "platform" {
		t.Fatalf("expected scopeType=platform, got %q", parsed.ScopeType)
	}
	if parsed.PlatformId != 1 {
		t.Fatalf("expected platformId=1, got %d", parsed.PlatformId)
	}
}
