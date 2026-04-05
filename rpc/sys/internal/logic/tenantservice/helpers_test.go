package tenantservicelogic

import (
	"regexp"
	"testing"
	"time"

	"github.com/feihua/zero-admin/pkg/scope"
	"github.com/feihua/zero-admin/rpc/sys/internal/tenantmodel"
)

func TestGenerateTenantCode(t *testing.T) {
	code := generateTenantCode(time.Date(2026, 3, 20, 12, 30, 45, 987654321, time.Local))
	if ok, _ := regexp.MatchString(`^TEN\d{18}$`, code); !ok {
		t.Fatalf("unexpected tenant code: %s", code)
	}
}

func TestNormalizeStringList(t *testing.T) {
	result := normalizeStringList([]string{" app ", "", "app", "mini-program", " mini-program "})
	if len(result) != 2 {
		t.Fatalf("unexpected normalized length: %v", result)
	}

	if result[0] != "app" || result[1] != "mini-program" {
		t.Fatalf("unexpected normalized values: %#v", result)
	}
}

func TestNormalizeChannelList(t *testing.T) {
	result := normalizeChannelList([]string{" app ", "", "app", "mini-program", " mini_program ", "pos"})
	if len(result) != 3 {
		t.Fatalf("unexpected normalized length: %v", result)
	}

	if result[0] != "app" || result[1] != "mini_program" || result[2] != "pos" {
		t.Fatalf("unexpected normalized channel values: %#v", result)
	}
}

func TestActivationStatusByTenantStatus(t *testing.T) {
	if got := activationStatusByTenantStatus(tenantmodel.TenantStatusPendingActivation); got != scope.ActivationStatusPending {
		t.Fatalf("unexpected pending activation status: %s", got)
	}

	if got := activationStatusByTenantStatus(tenantmodel.TenantStatusEnabled); got != scope.ActivationStatusActive {
		t.Fatalf("unexpected active status: %s", got)
	}
}

func TestValidateStatusTransition(t *testing.T) {
	if err := validateStatusTransition(tenantmodel.TenantStatusPendingActivation, tenantmodel.TenantStatusEnabled); err != nil {
		t.Fatalf("expected pending -> enabled to pass, got %v", err)
	}

	if err := validateStatusTransition(tenantmodel.TenantStatusArchived, tenantmodel.TenantStatusEnabled); err == nil {
		t.Fatal("expected archived -> enabled to fail")
	}
}
