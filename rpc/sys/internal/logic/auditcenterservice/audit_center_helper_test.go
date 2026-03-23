package auditcenterservicelogic

import (
	"testing"
	"time"

	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	"github.com/feihua/zero-admin/rpc/sys/sysclient"
)

func TestParseWindowRejectTooLongRange(t *testing.T) {
	_, err := parseWindow("2025-01-01 00:00:00", "2025-08-01 00:00:01")
	if err == nil {
		t.Fatal("expected range validation error")
	}
}

func TestAllowAuditItemHonorsTenantScope(t *testing.T) {
	current, _ := pkgscope.NormalizeGovernanceScope(pkgscope.SubjectTypeTenant, 1, 10, 0)
	item := auditCenterItem{PlatformID: 1, TenantID: 10, HappenedAt: time.Now()}
	if !allowAuditItem(item, current, &sysclient.QueryAuditCenterListReq{}) {
		t.Fatal("expected item in same scope to pass")
	}
	item.TenantID = 11
	if allowAuditItem(item, current, &sysclient.QueryAuditCenterListReq{}) {
		t.Fatal("expected item in different scope to be blocked")
	}
}
