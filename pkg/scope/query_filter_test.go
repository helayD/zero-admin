package scope

import "testing"

func TestScopeFilterSQLWithoutAlias(t *testing.T) {
	current, err := NormalizeGovernanceScope(SubjectTypeMerchant, 1, 12, 34)
	if err != nil {
		t.Fatalf("NormalizeGovernanceScope returned error: %v", err)
	}

	sql, args := ScopeFilterSQL("", current)
	if sql != "platform_id = ? AND tenant_id = ? AND merchant_id = ?" {
		t.Fatalf("unexpected sql: %s", sql)
	}
	if len(args) != 3 || args[0] != int64(1) || args[1] != int64(12) || args[2] != int64(34) {
		t.Fatalf("unexpected args: %#v", args)
	}
}

func TestScopeFilterSQLWithAlias(t *testing.T) {
	current, err := NormalizeGovernanceScope(SubjectTypeTenant, 1, 88, 0)
	if err != nil {
		t.Fatalf("NormalizeGovernanceScope returned error: %v", err)
	}

	sql, args := ScopeFilterSQL("spu", current)
	if sql != "spu.platform_id = ? AND spu.tenant_id = ?" {
		t.Fatalf("unexpected sql: %s", sql)
	}
	if len(args) != 2 || args[0] != int64(1) || args[1] != int64(88) {
		t.Fatalf("unexpected args: %#v", args)
	}
}

func TestScopeFilterSQLPlatformScopeOnlyBindsPlatformID(t *testing.T) {
	current, err := NormalizeGovernanceScope(SubjectTypePlatform, 1, 0, 0)
	if err != nil {
		t.Fatalf("NormalizeGovernanceScope returned error: %v", err)
	}

	sql, args := ScopeFilterSQL("o", current)
	if sql != "o.platform_id = ?" {
		t.Fatalf("unexpected sql: %s", sql)
	}
	if len(args) != 1 || args[0] != int64(1) {
		t.Fatalf("unexpected args: %#v", args)
	}
}

func TestEnsureScopeMatch(t *testing.T) {
	current, err := NormalizeGovernanceScope(SubjectTypeTenant, 1, 66, 0)
	if err != nil {
		t.Fatalf("NormalizeGovernanceScope returned error: %v", err)
	}

	if err = EnsureScopeMatch(current, 1, 66, 0, "mismatch"); err != nil {
		t.Fatalf("EnsureScopeMatch returned error: %v", err)
	}
	if err = EnsureScopeMatch(current, 1, 99, 0, "mismatch"); err == nil {
		t.Fatalf("expected mismatch error")
	}
}
