package common

import (
	"context"
	"strings"
	"testing"

	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newCatalogGuardTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}

	stmts := []string{
		"CREATE TABLE pms_product_spu (id INTEGER PRIMARY KEY, brand_id INTEGER, category_id INTEGER, is_deleted INTEGER DEFAULT 0)",
		"CREATE TABLE pms_product_category (id INTEGER PRIMARY KEY, parent_id INTEGER, is_deleted INTEGER DEFAULT 0)",
		"CREATE TABLE pms_product_attribute_group (id INTEGER PRIMARY KEY, category_id INTEGER, is_deleted INTEGER DEFAULT 0)",
		"CREATE TABLE pms_product_spec (id INTEGER PRIMARY KEY, category_id INTEGER, is_deleted INTEGER DEFAULT 0)",
		"CREATE TABLE pms_product_category_attribute_relation (id INTEGER PRIMARY KEY, product_category_id INTEGER, product_attribute_id INTEGER)",
		"CREATE TABLE pms_product_attribute (id INTEGER PRIMARY KEY, status INTEGER, group_id INTEGER, platform_id INTEGER, tenant_id INTEGER, merchant_id INTEGER, is_deleted INTEGER DEFAULT 0)",
		"CREATE TABLE pms_product_attribute_value (id INTEGER PRIMARY KEY, attribute_id INTEGER, is_deleted INTEGER DEFAULT 0)",
		"CREATE TABLE pms_product_spec_value (id INTEGER PRIMARY KEY, spec_id INTEGER, is_deleted INTEGER DEFAULT 0)",
	}

	for _, stmt := range stmts {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("exec %q failed: %v", stmt, err)
		}
	}

	return db
}

func TestEnsureCatalogDeleteAllowedRejectsReferences(t *testing.T) {
	db := newCatalogGuardTestDB(t)
	ctx := context.Background()

	checks := []struct {
		name    string
		seedSQL string
		check   func() error
		want    string
	}{
		{
			name:    "brand referenced by spu",
			seedSQL: "INSERT INTO pms_product_spu(id, brand_id, is_deleted) VALUES (1, 101, 0)",
			check:   func() error { return EnsureBrandDeleteAllowed(ctx, db, []int64{101}) },
			want:    "商品品牌已被商品SPU引用",
		},
		{
			name:    "category referenced by child category",
			seedSQL: "INSERT INTO pms_product_category(id, parent_id, is_deleted) VALUES (2, 201, 0)",
			check:   func() error { return EnsureCategoryDeleteAllowed(ctx, db, []int64{201}) },
			want:    "商品分类下仍存在子分类",
		},
		{
			name:    "attribute referenced by category relation",
			seedSQL: "INSERT INTO pms_product_category_attribute_relation(id, product_category_id, product_attribute_id) VALUES (1, 1, 301)",
			check:   func() error { return EnsureAttributeDeleteAllowed(ctx, db, []int64{301}) },
			want:    "商品属性已绑定商品分类",
		},
		{
			name:    "attribute group referenced by attribute",
			seedSQL: "INSERT INTO pms_product_attribute(id, group_id, status, platform_id, tenant_id, merchant_id, is_deleted) VALUES (1, 401, 1, 1, 10, 20, 0)",
			check:   func() error { return EnsureAttributeGroupDeleteAllowed(ctx, db, []int64{401}) },
			want:    "商品属性分组下仍存在商品属性",
		},
		{
			name:    "spec referenced by spec value",
			seedSQL: "INSERT INTO pms_product_spec_value(id, spec_id, is_deleted) VALUES (1, 501, 0)",
			check:   func() error { return EnsureSpecDeleteAllowed(ctx, db, []int64{501}) },
			want:    "商品规格下仍存在规格值",
		},
	}

	for _, tc := range checks {
		t.Run(tc.name, func(t *testing.T) {
			if err := db.Exec(tc.seedSQL).Error; err != nil {
				t.Fatalf("seed failed: %v", err)
			}
			err := tc.check()
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("expected error containing %q, got %v", tc.want, err)
			}
		})
	}
}

func TestEnsureCategoryAttributeBindingsValidatesExistenceScopeAndStatus(t *testing.T) {
	db := newCatalogGuardTestDB(t)
	ctx := context.Background()
	current, _ := pkgscope.NormalizeGovernanceScope(pkgscope.SubjectTypeMerchant, 1, 10, 20)
	other, _ := pkgscope.NormalizeGovernanceScope(pkgscope.SubjectTypeMerchant, 1, 10, 21)

	if err := db.Exec("INSERT INTO pms_product_attribute(id, status, group_id, platform_id, tenant_id, merchant_id, is_deleted) VALUES (1, 1, 11, 1, 10, 20, 0), (2, 0, 11, 1, 10, 20, 0), (3, 1, 11, ?, ?, ?, 0)", other.PlatformID, other.TenantID, other.MerchantID).Error; err != nil {
		t.Fatalf("seed attributes failed: %v", err)
	}

	tests := []struct {
		name string
		ids  []int64
		want string
	}{
		{name: "missing attribute", ids: []int64{99}, want: "不存在"},
		{name: "disabled attribute", ids: []int64{2}, want: "当前状态不可绑定"},
		{name: "cross scope attribute", ids: []int64{3}, want: "scope 不一致"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := EnsureCategoryAttributeBindings(ctx, db, current, tc.ids)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("expected error containing %q, got %v", tc.want, err)
			}
		})
	}

	if err := EnsureCategoryAttributeBindings(ctx, db, current, []int64{1, 1}); err != nil {
		t.Fatalf("expected valid binding to pass, got %v", err)
	}
}
