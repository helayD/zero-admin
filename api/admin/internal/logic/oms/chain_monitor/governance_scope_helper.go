package chain_monitor

import (
	"context"
	"strings"

	admincommon "github.com/feihua/zero-admin/api/admin/internal/common"
	"github.com/feihua/zero-admin/api/admin/internal/common/errorx"
	pkgscope "github.com/feihua/zero-admin/pkg/scope"
)

func resolveChainWriteScope(ctx context.Context, requested admincommon.RequestedGovernanceScope) (pkgscope.GovernanceScope, error) {
	current, err := admincommon.CurrentGovernanceScope(ctx)
	if err != nil {
		return pkgscope.GovernanceScope{}, errorx.NewDefaultError(err.Error())
	}

	if strings.TrimSpace(requested.ScopeType) == "" &&
		requested.PlatformID == 0 &&
		requested.TenantID == 0 &&
		requested.MerchantID == 0 {
		return current, nil
	}

	scopeType := strings.TrimSpace(requested.ScopeType)
	if scopeType == "" {
		scopeType = current.ScopeType
	}
	platformID := current.PlatformID
	if requested.PlatformID > 0 {
		platformID = requested.PlatformID
	}
	tenantID := current.TenantID
	if requested.TenantID > 0 {
		tenantID = requested.TenantID
	}
	merchantID := current.MerchantID
	if requested.MerchantID > 0 {
		merchantID = requested.MerchantID
	}

	resolved, err := pkgscope.NormalizeGovernanceScope(scopeType, platformID, tenantID, merchantID)
	if err != nil {
		return pkgscope.GovernanceScope{}, errorx.NewDefaultError(err.Error())
	}
	if !current.SameScope(resolved) {
		return pkgscope.GovernanceScope{}, errorx.NewCodeError("403", "当前主体不允许跨主体执行链路干预")
	}
	return resolved, nil
}
