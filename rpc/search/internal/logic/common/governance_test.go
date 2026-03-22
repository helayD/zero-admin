package common

import (
	"testing"

	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	"github.com/feihua/zero-admin/rpc/search/search"
)

func TestNormalizeProtoScopeDefaultsToPlatform(t *testing.T) {
	current, err := NormalizeProtoScope(nil)
	if err != nil {
		t.Fatalf("NormalizeProtoScope returned error: %v", err)
	}

	expected, err := pkgscope.NormalizeGovernanceScope(pkgscope.SubjectTypePlatform, pkgscope.DefaultPlatformID, 0, 0)
	if err != nil {
		t.Fatalf("NormalizeGovernanceScope returned error: %v", err)
	}
	if !current.SameScope(expected) {
		t.Fatalf("unexpected scope: got %+v want %+v", current, expected)
	}
}

func TestScopeFiltersIncludeAllScopeDimensions(t *testing.T) {
	filters := ScopeFilters(pkgscope.SubjectTypeMerchant, 1, 22, 33)
	if len(filters) != 4 {
		t.Fatalf("unexpected filter count: %d", len(filters))
	}
}

func TestProtoScopeCarriesNormalizedLabel(t *testing.T) {
	current, err := pkgscope.NormalizeGovernanceScope(pkgscope.SubjectTypeTenant, 1, 88, 0)
	if err != nil {
		t.Fatalf("NormalizeGovernanceScope returned error: %v", err)
	}

	proto := ProtoScope(current)
	if proto.ScopeType != pkgscope.SubjectTypeTenant || proto.TenantId != 88 {
		t.Fatalf("unexpected proto scope: %+v", proto)
	}
	if proto.ScopeLabel == "" {
		t.Fatalf("expected scope label to be populated")
	}
}

func TestKeywordMustUsesMatchAllWhenKeywordEmpty(t *testing.T) {
	must := KeywordMust(" ")
	if len(must) != 1 {
		t.Fatalf("unexpected must size: %d", len(must))
	}
}

func TestNormalizeProtoScopeRejectsInvalidMerchantScope(t *testing.T) {
	_, err := NormalizeProtoScope(&search.GovernanceScope{
		ScopeType: pkgscope.SubjectTypeMerchant,
	})
	if err == nil {
		t.Fatalf("expected invalid merchant scope to fail")
	}
}
