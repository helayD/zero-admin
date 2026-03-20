package scope

import (
	"errors"
	"fmt"
)

// ValidateRoleScope 验证角色创建/更新时的作用域合法性
func ValidateRoleScope(scopeType string, platformID, tenantID, merchantID int64) error {
	if scopeType == "" {
		return errors.New("角色作用域类型不能为空")
	}

	if platformID <= 0 {
		platformID = DefaultPlatformID
	}

	switch scopeType {
	case SubjectTypePlatform:
		if tenantID != 0 || merchantID != 0 {
			return errors.New("平台级角色不能绑定租户或商户")
		}
	case SubjectTypeTenant:
		if tenantID <= 0 {
			return errors.New("租户级角色必须指定有效的租户ID")
		}
		if merchantID != 0 {
			return errors.New("租户级角色不能绑定商户")
		}
	case SubjectTypeMerchant:
		if merchantID <= 0 {
			return errors.New("商户级角色必须指定有效的商户ID")
		}
	default:
		return fmt.Errorf("不支持的角色作用域类型：%s", scopeType)
	}

	return nil
}

// ValidateUserRoleAssignment 验证用户-角色分配的主体范围一致性
// userScope 是用户的主体范围，roleScope 是角色的主体范围
// 要求：用户主体范围 ⊆ 角色主体范围
func ValidateUserRoleAssignment(userScope, roleScope GovernanceScope) error {
	if userScope.ScopeType != roleScope.ScopeType {
		return fmt.Errorf("用户主体类型(%s)与角色主体类型(%s)不匹配", userScope.ScopeType, roleScope.ScopeType)
	}

	switch roleScope.ScopeType {
	case SubjectTypePlatform:
		// 平台角色可以分配给平台用户
		return nil
	case SubjectTypeTenant:
		if userScope.TenantID != roleScope.TenantID {
			return fmt.Errorf("用户所属租户(%d)与角色所属租户(%d)不匹配", userScope.TenantID, roleScope.TenantID)
		}
		return nil
	case SubjectTypeMerchant:
		if userScope.MerchantID != roleScope.MerchantID {
			return fmt.Errorf("用户所属商户(%d)与角色所属商户(%d)不匹配", userScope.MerchantID, roleScope.MerchantID)
		}
		return nil
	default:
		return fmt.Errorf("不支持的作用域类型：%s", roleScope.ScopeType)
	}
}

// RoleScopeToGovernanceScope 将角色的 scope 字段转换为 GovernanceScope 结构
func RoleScopeToGovernanceScope(scopeType string, platformID, tenantID, merchantID int64) GovernanceScope {
	if platformID <= 0 {
		platformID = DefaultPlatformID
	}
	return GovernanceScope{
		ScopeType:  scopeType,
		PlatformID: platformID,
		TenantID:   tenantID,
		MerchantID: merchantID,
	}
}

// BuildRoleScopeFilter 根据当前用户的作用域构建角色查询的过滤条件
// 返回 scopeType, tenantID, merchantID 用于 WHERE 条件
func BuildRoleScopeFilter(userScope GovernanceScope) (scopeType string, tenantID, merchantID int64) {
	return userScope.ScopeType, userScope.TenantID, userScope.MerchantID
}
