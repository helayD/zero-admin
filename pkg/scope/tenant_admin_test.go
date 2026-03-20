package scope

import (
	"encoding/json"
	"testing"
)

func TestBuildTenantAdminMetadata(t *testing.T) {
	raw, err := BuildTenantAdminMetadata(1, 12, "TEN202603200001", ActivationStatusPending)
	if err != nil {
		t.Fatalf("BuildTenantAdminMetadata() error = %v", err)
	}

	var payload TenantAdminMetadata
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if payload.SubjectType != SubjectTypeTenant {
		t.Fatalf("unexpected subjectType: %s", payload.SubjectType)
	}

	if payload.PlatformID != 1 || payload.TenantID != 12 {
		t.Fatalf("unexpected scope ids: %+v", payload)
	}

	if payload.TenantCode != "TEN202603200001" {
		t.Fatalf("unexpected tenantCode: %s", payload.TenantCode)
	}

	if payload.ActivationStatus != ActivationStatusPending {
		t.Fatalf("unexpected activation status: %s", payload.ActivationStatus)
	}

	if payload.RoleMode != RoleModeBootstrap {
		t.Fatalf("unexpected roleMode: %s", payload.RoleMode)
	}
}
