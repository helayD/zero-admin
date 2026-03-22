package common

import (
	"context"
	"path/filepath"
	"testing"

	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newOMSWriteScopeDB(t *testing.T) *gorm.DB {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "oms-write-scope.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}

	stmts := []string{
		`CREATE TABLE oms_order_main (
			id INTEGER PRIMARY KEY,
			platform_id INTEGER NOT NULL,
			tenant_id INTEGER NOT NULL,
			merchant_id INTEGER NOT NULL
		)`,
		`CREATE TABLE sys_security_event (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			trace_id TEXT,
			event_type TEXT,
			action TEXT,
			resource_type TEXT,
			resource_id INTEGER,
			scope_type TEXT,
			platform_id INTEGER,
			tenant_id INTEGER,
			merchant_id INTEGER,
			operator_id INTEGER,
			operator_name TEXT,
			request_summary TEXT,
			result TEXT,
			payload TEXT,
			created_at DATETIME
		)`,
	}
	for _, stmt := range stmts {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("exec schema failed: %v", err)
		}
	}

	return db
}

func TestEnsureOrderScopeRejectsCrossTenantBatch(t *testing.T) {
	db := newOMSWriteScopeDB(t)
	ctx := context.Background()
	current, err := pkgscope.NormalizeGovernanceScope(pkgscope.SubjectTypeTenant, pkgscope.DefaultPlatformID, 10, 0)
	if err != nil {
		t.Fatalf("normalize scope failed: %v", err)
	}

	if err := db.Exec(`INSERT INTO oms_order_main (id, platform_id, tenant_id, merchant_id) VALUES (1,1,10,0),(2,1,20,0)`).Error; err != nil {
		t.Fatalf("seed orders failed: %v", err)
	}

	_, err = EnsureOrderScope(ctx, db, current, []int64{1, 2}, "oms.order.close", 101, "tester", "close order")
	if err == nil {
		t.Fatal("expected cross-tenant order batch to be rejected")
	}

	var count int64
	if err := db.Table("sys_security_event").Count(&count).Error; err != nil {
		t.Fatalf("count security events failed: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 security event, got %d", count)
	}
}
