package common

import (
	"context"
	"fmt"
	"testing"

	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newWriteScopeTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}
	if err := db.Exec(`
		CREATE TABLE sys_user (
			id INTEGER PRIMARY KEY,
			platform_id INTEGER NOT NULL,
			tenant_id INTEGER NOT NULL,
			merchant_id INTEGER NOT NULL
		)
	`).Error; err != nil {
		t.Fatalf("create sys_user failed: %v", err)
	}
	if err := db.Exec(`
		INSERT INTO sys_user (id, platform_id, tenant_id, merchant_id) VALUES
		(1001, 1, 10, 0),
		(1002, 1, 10, 88)
	`).Error; err != nil {
		t.Fatalf("seed sys_user failed: %v", err)
	}

	return db
}

func TestResolveWriteScopeUsesActorScopeByDefault(t *testing.T) {
	db := newWriteScopeTestDB(t)

	scope, err := ResolveWriteScope(context.Background(), db, nil, 1002)
	if err != nil {
		t.Fatalf("ResolveWriteScope returned error: %v", err)
	}
	if scope.ScopeType != pkgscope.SubjectTypeMerchant || scope.TenantID != 10 || scope.MerchantID != 88 {
		t.Fatalf("unexpected resolved scope: %+v", scope)
	}
}

func TestResolveWriteScopeRejectsTenantExpansion(t *testing.T) {
	db := newWriteScopeTestDB(t)

	_, err := ResolveWriteScope(context.Background(), db, &smsclient.GovernanceScope{
		ScopeType:  pkgscope.SubjectTypePlatform,
		PlatformId: 1,
	}, 1001)
	if err == nil || err.Error() != "当前主体不允许切换写入范围" {
		t.Fatalf("expected tenant expansion error, got %v", err)
	}
}

func TestResolveWriteScopeAllowsPlatformActorOverride(t *testing.T) {
	db := newWriteScopeTestDB(t)

	scope, err := ResolveWriteScope(context.Background(), db, &smsclient.GovernanceScope{
		ScopeType:  pkgscope.SubjectTypeMerchant,
		PlatformId: 1,
		TenantId:   10,
		MerchantId: 88,
	}, 0)
	if err != nil {
		t.Fatalf("ResolveWriteScope returned error: %v", err)
	}
	if scope.ScopeType != pkgscope.SubjectTypeMerchant || scope.MerchantID != 88 {
		t.Fatalf("unexpected resolved scope: %+v", scope)
	}
}
