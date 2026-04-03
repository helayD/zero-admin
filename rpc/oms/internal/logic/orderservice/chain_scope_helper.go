package orderservicelogic

import (
	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	"github.com/feihua/zero-admin/rpc/oms/gen/model"
)

func hasRequiredChainScope(platformID int64) bool {
	return platformID > 0
}

func buildOrderGovernanceScope(order *model.OmsOrderMain) pkgscope.GovernanceScope {
	scopeType := pkgscope.SubjectTypePlatform
	if order.MerchantID > 0 {
		scopeType = pkgscope.SubjectTypeMerchant
	} else if order.TenantID > 0 {
		scopeType = pkgscope.SubjectTypeTenant
	}

	return pkgscope.GovernanceScope{
		ScopeType:  scopeType,
		PlatformID: order.PlatformID,
		TenantID:   order.TenantID,
		MerchantID: order.MerchantID,
	}
}
