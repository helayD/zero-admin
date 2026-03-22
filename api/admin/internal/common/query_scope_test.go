package common

import (
	"context"
	"encoding/json"
	"testing"

	pkgscope "github.com/feihua/zero-admin/pkg/scope"
)

func TestResolveQueryGovernanceScopeUsesCurrentScopeByDefault(t *testing.T) {
	ctx := newScopeContext(pkgscope.SubjectTypeTenant, pkgscope.DefaultPlatformID, 88, 0)

	resolved, err := ResolveQueryGovernanceScope(ctx, RequestedGovernanceScope{})
	if err != nil {
		t.Fatalf("ResolveQueryGovernanceScope returned error: %v", err)
	}

	assertSameScope(t, resolved, pkgscope.SubjectTypeTenant, pkgscope.DefaultPlatformID, 88, 0)
}

func TestResolveQueryGovernanceScopeAllowsPlatformToSwitch(t *testing.T) {
	ctx := newScopeContext(pkgscope.SubjectTypePlatform, pkgscope.DefaultPlatformID, 0, 0)

	resolved, err := ResolveQueryGovernanceScope(ctx, RequestedGovernanceScope{
		ScopeType:  pkgscope.SubjectTypeMerchant,
		MerchantID: 3001,
	})
	if err != nil {
		t.Fatalf("ResolveQueryGovernanceScope returned error: %v", err)
	}

	assertSameScope(t, resolved, pkgscope.SubjectTypeMerchant, pkgscope.DefaultPlatformID, 0, 3001)
}

func TestResolveQueryGovernanceScopeRejectsTenantExpansion(t *testing.T) {
	ctx := newScopeContext(pkgscope.SubjectTypeTenant, pkgscope.DefaultPlatformID, 88, 0)

	_, err := ResolveQueryGovernanceScope(ctx, RequestedGovernanceScope{
		ScopeType: pkgscope.SubjectTypePlatform,
	})
	if err == nil {
		t.Fatal("expected tenant scope expansion to be rejected")
	}
}

func TestResolveQueryGovernanceScopeRejectsMerchantSwitch(t *testing.T) {
	ctx := newScopeContext(pkgscope.SubjectTypeMerchant, pkgscope.DefaultPlatformID, 88, 3001)

	_, err := ResolveQueryGovernanceScope(ctx, RequestedGovernanceScope{
		ScopeType:  pkgscope.SubjectTypeMerchant,
		MerchantID: 3002,
	})
	if err == nil {
		t.Fatal("expected merchant scope switch to be rejected")
	}
}

func TestCurrentGovernanceScopeRejectsMissingScope(t *testing.T) {
	_, err := CurrentGovernanceScope(context.Background())
	if err == nil {
		t.Fatal("expected missing scope in context to fail")
	}
}

func newScopeContext(scopeType string, platformID, tenantID, merchantID int64) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, "scopeType", scopeType)
	ctx = context.WithValue(ctx, "platformId", json.Number("1"))
	if platformID != pkgscope.DefaultPlatformID {
		ctx = context.WithValue(ctx, "platformId", platformID)
	}
	ctx = context.WithValue(ctx, "tenantId", tenantID)
	ctx = context.WithValue(ctx, "merchantId", merchantID)
	return ctx
}

func assertSameScope(t *testing.T, current pkgscope.GovernanceScope, scopeType string, platformID, tenantID, merchantID int64) {
	t.Helper()

	expected, err := pkgscope.NormalizeGovernanceScope(scopeType, platformID, tenantID, merchantID)
	if err != nil {
		t.Fatalf("NormalizeGovernanceScope returned error: %v", err)
	}
	if !current.SameScope(expected) {
		t.Fatalf("unexpected scope: got %+v want %+v", current, expected)
	}
}
