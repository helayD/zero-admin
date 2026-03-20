package audit

import (
	"encoding/json"
	"testing"
)

func TestEncodeTenantPayload(t *testing.T) {
	raw, err := EncodeTenantPayload(TenantPayload{
		TenantID:              9,
		TenantCode:            "TEN202603200001",
		TenantName:            "华东运营租户",
		Action:                ActionTenantCreated,
		CurrentStatus:         2,
		AvailableChannels:     []string{"app", "mini-program"},
		FeatureFlags:          []string{"audit", "merchant-onboarding"},
		PrimaryAdminUserID:    18,
		PrimaryAdminUserName:  "tenant_admin_hd",
		PrimaryAdminMobile:    "13800138000",
		AdminActivationStatus: "pending-activation",
	})
	if err != nil {
		t.Fatalf("EncodeTenantPayload() error = %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if payload["tenantId"] != float64(9) {
		t.Fatalf("unexpected tenantId: %v", payload["tenantId"])
	}

	if payload["action"] != ActionTenantCreated {
		t.Fatalf("unexpected action: %v", payload["action"])
	}
}
