package scope

import (
	"testing"
)

func TestValidateRoleScope_Platform(t *testing.T) {
	// 平台级角色，tenantID=0, merchantID=0 应通过
	if err := ValidateRoleScope(SubjectTypePlatform, 1, 0, 0); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}

	// 平台级角色绑定租户应失败
	if err := ValidateRoleScope(SubjectTypePlatform, 1, 10, 0); err == nil {
		t.Fatal("expected error for platform role with tenantID")
	}

	// 平台级角色绑定商户应失败
	if err := ValidateRoleScope(SubjectTypePlatform, 1, 0, 5); err == nil {
		t.Fatal("expected error for platform role with merchantID")
	}
}

func TestValidateRoleScope_Tenant(t *testing.T) {
	// 租户级角色，tenantID>0, merchantID=0 应通过
	if err := ValidateRoleScope(SubjectTypeTenant, 1, 10, 0); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}

	// 租户级角色缺少tenantID应失败
	if err := ValidateRoleScope(SubjectTypeTenant, 1, 0, 0); err == nil {
		t.Fatal("expected error for tenant role without tenantID")
	}

	// 租户级角色绑定商户应失败
	if err := ValidateRoleScope(SubjectTypeTenant, 1, 10, 5); err == nil {
		t.Fatal("expected error for tenant role with merchantID")
	}
}

func TestValidateRoleScope_Merchant(t *testing.T) {
	// 商户级角色，merchantID>0 应通过
	if err := ValidateRoleScope(SubjectTypeMerchant, 1, 0, 5); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}

	// 商户级角色缺少merchantID应失败
	if err := ValidateRoleScope(SubjectTypeMerchant, 1, 0, 0); err == nil {
		t.Fatal("expected error for merchant role without merchantID")
	}
}

func TestValidateRoleScope_EmptyType(t *testing.T) {
	if err := ValidateRoleScope("", 1, 0, 0); err == nil {
		t.Fatal("expected error for empty scopeType")
	}
}

func TestValidateRoleScope_UnknownType(t *testing.T) {
	if err := ValidateRoleScope("unknown", 1, 0, 0); err == nil {
		t.Fatal("expected error for unknown scopeType")
	}
}

func TestValidateUserRoleAssignment(t *testing.T) {
	// 同租户分配应通过
	userScope := GovernanceScope{ScopeType: SubjectTypeTenant, PlatformID: 1, TenantID: 10}
	roleScope := GovernanceScope{ScopeType: SubjectTypeTenant, PlatformID: 1, TenantID: 10}
	if err := ValidateUserRoleAssignment(userScope, roleScope); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}

	// 不同租户分配应失败
	userScope2 := GovernanceScope{ScopeType: SubjectTypeTenant, PlatformID: 1, TenantID: 10}
	roleScope2 := GovernanceScope{ScopeType: SubjectTypeTenant, PlatformID: 1, TenantID: 20}
	if err := ValidateUserRoleAssignment(userScope2, roleScope2); err == nil {
		t.Fatal("expected error for mismatched tenantID")
	}

	// 不同 scopeType 分配应失败
	userScope3 := GovernanceScope{ScopeType: SubjectTypePlatform, PlatformID: 1}
	roleScope3 := GovernanceScope{ScopeType: SubjectTypeTenant, PlatformID: 1, TenantID: 10}
	if err := ValidateUserRoleAssignment(userScope3, roleScope3); err == nil {
		t.Fatal("expected error for mismatched scopeType")
	}

	// 平台级分配应通过
	userScope4 := GovernanceScope{ScopeType: SubjectTypePlatform, PlatformID: 1}
	roleScope4 := GovernanceScope{ScopeType: SubjectTypePlatform, PlatformID: 1}
	if err := ValidateUserRoleAssignment(userScope4, roleScope4); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}

	// 同商户分配应通过
	userScope5 := GovernanceScope{ScopeType: SubjectTypeMerchant, PlatformID: 1, MerchantID: 5}
	roleScope5 := GovernanceScope{ScopeType: SubjectTypeMerchant, PlatformID: 1, MerchantID: 5}
	if err := ValidateUserRoleAssignment(userScope5, roleScope5); err != nil {
		t.Fatalf("expected nil, got %v", err)
	}

	// 不同商户分配应失败
	userScope6 := GovernanceScope{ScopeType: SubjectTypeMerchant, PlatformID: 1, MerchantID: 5}
	roleScope6 := GovernanceScope{ScopeType: SubjectTypeMerchant, PlatformID: 1, MerchantID: 6}
	if err := ValidateUserRoleAssignment(userScope6, roleScope6); err == nil {
		t.Fatal("expected error for mismatched merchantID")
	}
}

func TestRoleScopeToGovernanceScope(t *testing.T) {
	gs := RoleScopeToGovernanceScope(SubjectTypeTenant, 1, 10, 0)
	if gs.ScopeType != SubjectTypeTenant {
		t.Fatalf("expected %s, got %s", SubjectTypeTenant, gs.ScopeType)
	}
	if gs.PlatformID != 1 {
		t.Fatalf("expected 1, got %d", gs.PlatformID)
	}
	if gs.TenantID != 10 {
		t.Fatalf("expected 10, got %d", gs.TenantID)
	}

	// platformID <= 0 应使用默认值
	gs2 := RoleScopeToGovernanceScope(SubjectTypePlatform, 0, 0, 0)
	if gs2.PlatformID != DefaultPlatformID {
		t.Fatalf("expected default platformID %d, got %d", DefaultPlatformID, gs2.PlatformID)
	}
}

func TestBuildRoleScopeFilter(t *testing.T) {
	userScope := GovernanceScope{ScopeType: SubjectTypeTenant, PlatformID: 1, TenantID: 10, MerchantID: 0}
	scopeType, tenantID, merchantID := BuildRoleScopeFilter(userScope)
	if scopeType != SubjectTypeTenant {
		t.Fatalf("expected %s, got %s", SubjectTypeTenant, scopeType)
	}
	if tenantID != 10 {
		t.Fatalf("expected 10, got %d", tenantID)
	}
	if merchantID != 0 {
		t.Fatalf("expected 0, got %d", merchantID)
	}
}
