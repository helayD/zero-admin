package common

import (
	"context"
	"path/filepath"
	"testing"

	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newPMSWriteScopeDB(t *testing.T) *gorm.DB {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "pms-write-scope.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}

	stmts := []string{
		`CREATE TABLE pms_product_spu (
			id INTEGER PRIMARY KEY,
			platform_id INTEGER NOT NULL,
			tenant_id INTEGER NOT NULL,
			merchant_id INTEGER NOT NULL
		)`,
		`CREATE TABLE pms_product_sku (
			id INTEGER PRIMARY KEY,
			spu_id INTEGER NOT NULL,
			platform_id INTEGER NOT NULL,
			tenant_id INTEGER NOT NULL,
			merchant_id INTEGER NOT NULL
		)`,
		`CREATE TABLE pms_product_attribute_value (
			id INTEGER PRIMARY KEY,
			spu_id INTEGER NOT NULL,
			platform_id INTEGER NOT NULL,
			tenant_id INTEGER NOT NULL,
			merchant_id INTEGER NOT NULL
		)`,
		`CREATE TABLE pms_member_price (
			id INTEGER PRIMARY KEY,
			product_id INTEGER NOT NULL,
			platform_id INTEGER NOT NULL,
			tenant_id INTEGER NOT NULL,
			merchant_id INTEGER NOT NULL
		)`,
		`CREATE TABLE pms_product_ladder (
			id INTEGER PRIMARY KEY,
			product_id INTEGER NOT NULL,
			platform_id INTEGER NOT NULL,
			tenant_id INTEGER NOT NULL,
			merchant_id INTEGER NOT NULL
		)`,
		`CREATE TABLE pms_product_full_reduction (
			id INTEGER PRIMARY KEY,
			product_id INTEGER NOT NULL,
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

func TestEnsureProductScopeRejectsMixedBatchAndRecordsSecurityEvent(t *testing.T) {
	db := newPMSWriteScopeDB(t)
	ctx := context.Background()
	current, err := pkgscope.NormalizeGovernanceScope(pkgscope.SubjectTypeTenant, pkgscope.DefaultPlatformID, 10, 0)
	if err != nil {
		t.Fatalf("normalize scope failed: %v", err)
	}

	if err := db.Exec(`INSERT INTO pms_product_spu (id, platform_id, tenant_id, merchant_id) VALUES (1,1,10,0),(2,1,20,0)`).Error; err != nil {
		t.Fatalf("seed spu failed: %v", err)
	}

	_, err = EnsureProductScope(ctx, db, current, []int64{1, 2}, "pms.product_spu.update", 100, "tester", "mixed batch")
	if err == nil {
		t.Fatal("expected mixed batch to be rejected")
	}

	var count int64
	if err := db.Table("sys_security_event").Count(&count).Error; err != nil {
		t.Fatalf("count security event failed: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 security event, got %d", count)
	}
}

func TestApplyProductScopePropagatesToSkuRows(t *testing.T) {
	db := newPMSWriteScopeDB(t)
	ctx := context.Background()
	current, err := pkgscope.NormalizeGovernanceScope(pkgscope.SubjectTypeMerchant, pkgscope.DefaultPlatformID, 10, 3001)
	if err != nil {
		t.Fatalf("normalize scope failed: %v", err)
	}

	if err := db.Exec(`INSERT INTO pms_product_spu (id, platform_id, tenant_id, merchant_id) VALUES (1,1,0,0)`).Error; err != nil {
		t.Fatalf("seed spu failed: %v", err)
	}
	if err := db.Exec(`INSERT INTO pms_product_sku (id, spu_id, platform_id, tenant_id, merchant_id) VALUES (11,1,1,0,0)`).Error; err != nil {
		t.Fatalf("seed sku failed: %v", err)
	}
	if err := db.Exec(`INSERT INTO pms_product_attribute_value (id, spu_id, platform_id, tenant_id, merchant_id) VALUES (21,1,1,0,0)`).Error; err != nil {
		t.Fatalf("seed attribute value failed: %v", err)
	}
	if err := db.Exec(`INSERT INTO pms_member_price (id, product_id, platform_id, tenant_id, merchant_id) VALUES (31,1,1,0,0)`).Error; err != nil {
		t.Fatalf("seed member price failed: %v", err)
	}
	if err := db.Exec(`INSERT INTO pms_product_ladder (id, product_id, platform_id, tenant_id, merchant_id) VALUES (41,1,1,0,0)`).Error; err != nil {
		t.Fatalf("seed ladder failed: %v", err)
	}
	if err := db.Exec(`INSERT INTO pms_product_full_reduction (id, product_id, platform_id, tenant_id, merchant_id) VALUES (51,1,1,0,0)`).Error; err != nil {
		t.Fatalf("seed full reduction failed: %v", err)
	}

	if err := ApplyProductScope(ctx, db, 1, current); err != nil {
		t.Fatalf("apply product scope failed: %v", err)
	}

	var spuRow struct {
		TenantID   int64 `gorm:"column:tenant_id"`
		MerchantID int64 `gorm:"column:merchant_id"`
	}
	if err := db.Table("pms_product_spu").Select("tenant_id, merchant_id").Where("id = 1").Take(&spuRow).Error; err != nil {
		t.Fatalf("query updated spu failed: %v", err)
	}
	if spuRow.TenantID != 10 || spuRow.MerchantID != 3001 {
		t.Fatalf("unexpected spu scope: %+v", spuRow)
	}

	var skuRow struct {
		TenantID   int64 `gorm:"column:tenant_id"`
		MerchantID int64 `gorm:"column:merchant_id"`
	}
	if err := db.Table("pms_product_sku").Select("tenant_id, merchant_id").Where("id = 11").Take(&skuRow).Error; err != nil {
		t.Fatalf("query updated sku failed: %v", err)
	}
	if skuRow.TenantID != 10 || skuRow.MerchantID != 3001 {
		t.Fatalf("unexpected sku scope: %+v", skuRow)
	}

	for _, target := range []struct {
		table string
		id    int64
		label string
	}{
		{table: "pms_product_attribute_value", id: 21, label: "attribute value"},
		{table: "pms_member_price", id: 31, label: "member price"},
		{table: "pms_product_ladder", id: 41, label: "product ladder"},
		{table: "pms_product_full_reduction", id: 51, label: "full reduction"},
	} {
		var row struct {
			TenantID   int64 `gorm:"column:tenant_id"`
			MerchantID int64 `gorm:"column:merchant_id"`
		}
		if err := db.Table(target.table).Select("tenant_id, merchant_id").Where("id = ?", target.id).Take(&row).Error; err != nil {
			t.Fatalf("query updated %s failed: %v", target.label, err)
		}
		if row.TenantID != 10 || row.MerchantID != 3001 {
			t.Fatalf("unexpected %s scope: %+v", target.label, row)
		}
	}
}
