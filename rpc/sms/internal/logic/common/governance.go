package common

import (
	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"
)

func NormalizeProtoScope(input *smsclient.GovernanceScope) (pkgscope.GovernanceScope, error) {
	if input == nil {
		return pkgscope.NormalizeGovernanceScope(pkgscope.SubjectTypePlatform, pkgscope.DefaultPlatformID, 0, 0)
	}

	return pkgscope.NormalizeGovernanceScope(input.ScopeType, input.PlatformId, input.TenantId, input.MerchantId)
}
