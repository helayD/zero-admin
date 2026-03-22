package common

import (
	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	"github.com/feihua/zero-admin/rpc/search/search"
)

func NormalizeProtoScope(input *search.GovernanceScope) (pkgscope.GovernanceScope, error) {
	if input == nil {
		return pkgscope.NormalizeGovernanceScope(pkgscope.SubjectTypePlatform, pkgscope.DefaultPlatformID, 0, 0)
	}

	return pkgscope.NormalizeGovernanceScope(input.ScopeType, input.PlatformId, input.TenantId, input.MerchantId)
}

func ProtoScope(current pkgscope.GovernanceScope) *search.GovernanceScope {
	return &search.GovernanceScope{
		ScopeType:  current.ScopeType,
		PlatformId: current.PlatformID,
		TenantId:   current.TenantID,
		MerchantId: current.MerchantID,
		ScopeLabel: current.Label(),
	}
}
