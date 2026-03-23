package productcategoryservicelogic

import (
	"context"
	"strings"
	"testing"

	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newCategoryRelationTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}

	stmts := []string{
		"CREATE TABLE pms_product_attribute (id INTEGER PRIMARY KEY, status INTEGER, platform_id INTEGER, tenant_id INTEGER, merchant_id INTEGER, is_deleted INTEGER DEFAULT 0)",
		"CREATE TABLE pms_product_category_attribute_relation (id INTEGER PRIMARY KEY AUTOINCREMENT, product_category_id INTEGER, product_attribute_id INTEGER, platform_id INTEGER, tenant_id INTEGER, merchant_id INTEGER)",
	}
	for _, stmt := range stmts {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("exec %q failed: %v", stmt, err)
		}
	}
	return db
}

func TestReplaceCategoryRelationsRejectsInvalidAttributeBindings(t *testing.T) {
	db := newCategoryRelationTestDB(t)
	current, _ := pkgscope.NormalizeGovernanceScope(pkgscope.SubjectTypeMerchant, 1, 10, 20)

	if err := db.Exec("INSERT INTO pms_product_attribute(id, status, platform_id, tenant_id, merchant_id, is_deleted) VALUES (1, 0, 1, 10, 20, 0), (2, 1, 1, 10, 21, 0)").Error; err != nil {
		t.Fatalf("seed attribute failed: %v", err)
	}

	err := db.WithContext(context.Background()).Transaction(func(tx *gorm.DB) error {
		return replaceCategoryRelations(tx, 1001, []int64{1}, current)
	})
	if err == nil || !strings.Contains(err.Error(), "只能绑定启用状态") {
		t.Fatalf("expected disabled attribute error, got %v", err)
	}

	err = db.WithContext(context.Background()).Transaction(func(tx *gorm.DB) error {
		return replaceCategoryRelations(tx, 1001, []int64{2}, current)
	})
	if err == nil || !strings.Contains(err.Error(), "作用域不一致") {
		t.Fatalf("expected cross scope error, got %v", err)
	}
}

func TestReplaceCategoryRelationsReplacesWithUniqueBindings(t *testing.T) {
	db := newCategoryRelationTestDB(t)
	current, _ := pkgscope.NormalizeGovernanceScope(pkgscope.SubjectTypeMerchant, 1, 10, 20)

	if err := db.Exec("INSERT INTO pms_product_attribute(id, status, platform_id, tenant_id, merchant_id, is_deleted) VALUES (1, 1, 1, 10, 20, 0), (2, 1, 1, 10, 20, 0)").Error; err != nil {
		t.Fatalf("seed attribute failed: %v", err)
	}
	if err := db.Exec("INSERT INTO pms_product_category_attribute_relation(product_category_id, product_attribute_id, platform_id, tenant_id, merchant_id) VALUES (1001, 99, 1, 10, 20)").Error; err != nil {
		t.Fatalf("seed relation failed: %v", err)
	}

	if err := db.WithContext(context.Background()).Transaction(func(tx *gorm.DB) error {
		return replaceCategoryRelations(tx, 1001, []int64{1, 1, 2}, current)
	}); err != nil {
		t.Fatalf("replace relations failed: %v", err)
	}

	var count int64
	if err := db.Table("pms_product_category_attribute_relation").Where("product_category_id = ?", 1001).Count(&count).Error; err != nil {
		t.Fatalf("count relation failed: %v", err)
	}
	if count != 2 {
		t.Fatalf("expected 2 unique relations, got %d", count)
	}
}
