package common

import (
	"github.com/feihua/zero-admin/pkg/scope"
	"github.com/feihua/zero-admin/rpc/sys/sysclient"
)

func BuildGovernanceScope(scopeType string, platformID, tenantID, merchantID int64) (*sysclient.GovernanceScope, error) {
	normalized, err := scope.NormalizeGovernanceScope(scopeType, platformID, tenantID, merchantID)
	if err != nil {
		return nil, err
	}

	return &sysclient.GovernanceScope{
		ScopeType:  normalized.ScopeType,
		PlatformId: normalized.PlatformID,
		TenantId:   normalized.TenantID,
		MerchantId: normalized.MerchantID,
		ScopeLabel: normalized.Label(),
	}, nil
}

func ReadGovernanceScope(item *sysclient.GovernanceScope) (scopeType, scopeLabel string, platformID, tenantID, merchantID int64) {
	if item == nil {
		return scope.SubjectTypePlatform, "平台级", scope.DefaultPlatformID, 0, 0
	}

	return item.ScopeType, item.ScopeLabel, item.PlatformId, item.TenantId, item.MerchantId
}
