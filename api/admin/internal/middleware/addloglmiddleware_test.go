package middleware

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	pkgscope "github.com/feihua/zero-admin/pkg/scope"
)

func TestBuildGovernanceOperateExtraIncludesResourceIDsAndScope(t *testing.T) {
	body := []byte(`{"ids":[12,13],"status":1}`)
	req := httptest.NewRequest(http.MethodPost, "/api/pms/product/updatePublishStatus", bytes.NewReader(body))
	scope, err := pkgscope.NormalizeGovernanceScope(pkgscope.SubjectTypeTenant, pkgscope.DefaultPlatformID, 88, 0)
	if err != nil {
		t.Fatalf("normalize scope failed: %v", err)
	}

	extra := buildGovernanceOperateExtra(req, body, []byte(`{"code":"000000"}`), scope)

	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(extra), &payload); err != nil {
		t.Fatalf("unmarshal extra failed: %v", err)
	}

	if payload["action"] != "updatePublishStatus" {
		t.Fatalf("unexpected action: %v", payload["action"])
	}
	if payload["resourceType"] != "product_spu" {
		t.Fatalf("unexpected resourceType: %v", payload["resourceType"])
	}
	if payload["scopeType"] != pkgscope.SubjectTypeTenant {
		t.Fatalf("unexpected scopeType: %v", payload["scopeType"])
	}

	resourceIDs, ok := payload["resourceIds"].([]interface{})
	if !ok || len(resourceIDs) != 2 {
		t.Fatalf("unexpected resourceIds: %#v", payload["resourceIds"])
	}
}

func TestReadGovernanceScopeContextFallsBackToPlatform(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/cms/subject/querySubjectList", nil)
	req = req.WithContext(context.Background())

	current := readGovernanceScopeContext(req)
	if current.ScopeType != pkgscope.SubjectTypePlatform {
		t.Fatalf("unexpected scope type: %s", current.ScopeType)
	}
	if current.PlatformID != pkgscope.DefaultPlatformID {
		t.Fatalf("unexpected platform id: %d", current.PlatformID)
	}
}
