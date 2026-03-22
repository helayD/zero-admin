package common

import (
	"testing"

	pkgscope "github.com/feihua/zero-admin/pkg/scope"
)

func TestResolveWriteGovernanceScopeUsesCurrentScopeByDefault(t *testing.T) {
	ctx := newScopeContext(pkgscope.SubjectTypeMerchant, pkgscope.DefaultPlatformID, 88, 3001)

	resolved, err := ResolveWriteGovernanceScope(ctx, RequestedGovernanceScope{})
	if err != nil {
		t.Fatalf("ResolveWriteGovernanceScope returned error: %v", err)
	}

	assertSameScope(t, resolved, pkgscope.SubjectTypeMerchant, pkgscope.DefaultPlatformID, 88, 3001)
}

func TestResolveWriteGovernanceScopeAllowsPlatformSwitch(t *testing.T) {
	ctx := newScopeContext(pkgscope.SubjectTypePlatform, pkgscope.DefaultPlatformID, 0, 0)

	resolved, err := ResolveWriteGovernanceScope(ctx, RequestedGovernanceScope{
		ScopeType:  pkgscope.SubjectTypeTenant,
		TenantID:   88,
		PlatformID: pkgscope.DefaultPlatformID,
	})
	if err != nil {
		t.Fatalf("ResolveWriteGovernanceScope returned error: %v", err)
	}

	assertSameScope(t, resolved, pkgscope.SubjectTypeTenant, pkgscope.DefaultPlatformID, 88, 0)
}

func TestResolveWriteGovernanceScopeRejectsTenantExpansion(t *testing.T) {
	ctx := newScopeContext(pkgscope.SubjectTypeTenant, pkgscope.DefaultPlatformID, 88, 0)

	_, err := ResolveWriteGovernanceScope(ctx, RequestedGovernanceScope{
		ScopeType:  pkgscope.SubjectTypeMerchant,
		TenantID:   99,
		MerchantID: 3001,
	})
	if err == nil {
		t.Fatal("expected tenant scope expansion to be rejected")
	}
}
