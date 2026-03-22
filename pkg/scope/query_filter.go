package scope

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
)

func ScopeFilterSQL(alias string, current GovernanceScope) (string, []interface{}) {
	prefix := alias
	if prefix != "" {
		prefix += "."
	}

	return fmt.Sprintf("%splatform_id = ? AND %stenant_id = ? AND %smerchant_id = ?", prefix, prefix, prefix), []interface{}{
		current.PlatformID,
		current.TenantID,
		current.MerchantID,
	}
}

func ApplyGovernanceScope(db *gorm.DB, current GovernanceScope, alias string) *gorm.DB {
	scopeWhere, scopeArgs := ScopeFilterSQL(alias, current)
	return db.Where(scopeWhere, scopeArgs...)
}

func EnsureScopeMatch(current GovernanceScope, platformID, tenantID, merchantID int64, message string) error {
	if !current.SameScope(DefaultScope(platformID, tenantID, merchantID)) {
		return errors.New(message)
	}

	return nil
}

func DefaultScope(platformID, tenantID, merchantID int64) GovernanceScope {
	normalized, _ := NormalizeGovernanceScope("", platformID, tenantID, merchantID)
	return normalized
}
