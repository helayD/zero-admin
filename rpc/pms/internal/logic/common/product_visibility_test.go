package common

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newProductVisibilityTestDB(t *testing.T) (*gorm.DB, pkgscope.GovernanceScope) {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "product-visibility.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}

	schema := []string{
		`CREATE TABLE pms_product_spu (
			id INTEGER PRIMARY KEY,
			name TEXT,
			category_id INTEGER,
			brand_id INTEGER,
			main_pic TEXT,
			publish_status INTEGER DEFAULT 0,
			verify_status INTEGER DEFAULT 0,
			recommend_status INTEGER DEFAULT 0,
			preview_status INTEGER DEFAULT 0,
			stock INTEGER DEFAULT 0,
			platform_id INTEGER DEFAULT 1,
			tenant_id INTEGER DEFAULT 0,
			merchant_id INTEGER DEFAULT 0,
			is_deleted INTEGER DEFAULT 0
		)`,
		`CREATE TABLE pms_product_sku (
			id INTEGER PRIMARY KEY,
			spu_id INTEGER,
			stock INTEGER DEFAULT 0,
			platform_id INTEGER DEFAULT 1,
			tenant_id INTEGER DEFAULT 0,
			merchant_id INTEGER DEFAULT 0,
			is_deleted INTEGER DEFAULT 0
		)`,
		`CREATE TABLE pms_product_category (
			id INTEGER PRIMARY KEY,
			is_enabled INTEGER DEFAULT 1,
			is_deleted INTEGER DEFAULT 0,
			platform_id INTEGER DEFAULT 1,
			tenant_id INTEGER DEFAULT 0,
			merchant_id INTEGER DEFAULT 0
		)`,
		`CREATE TABLE pms_product_brand (
			id INTEGER PRIMARY KEY,
			is_enabled INTEGER DEFAULT 1,
			is_deleted INTEGER DEFAULT 0,
			platform_id INTEGER DEFAULT 1,
			tenant_id INTEGER DEFAULT 0,
			merchant_id INTEGER DEFAULT 0
		)`,
		`CREATE TABLE sys_tenant (
			id INTEGER PRIMARY KEY,
			status INTEGER DEFAULT 1
		)`,
		`CREATE TABLE sys_merchant (
			id INTEGER PRIMARY KEY,
			business_status INTEGER DEFAULT 1
		)`,
	}
	for _, stmt := range schema {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("exec schema failed: %v", err)
		}
	}

	scope, err := pkgscope.NormalizeGovernanceScope(pkgscope.SubjectTypeMerchant, 1, 10, 301)
	if err != nil {
		t.Fatalf("normalize scope failed: %v", err)
	}

	seeds := []string{
		`INSERT INTO sys_tenant(id, status) VALUES (10, 1)`,
		`INSERT INTO sys_merchant(id, business_status) VALUES (301, 1)`,
		`INSERT INTO pms_product_category(id, is_enabled, is_deleted, platform_id, tenant_id, merchant_id) VALUES (11, 1, 0, 1, 10, 301)`,
		`INSERT INTO pms_product_brand(id, is_enabled, is_deleted, platform_id, tenant_id, merchant_id) VALUES (21, 1, 0, 1, 10, 301)`,
	}
	for _, stmt := range seeds {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("seed base rows failed: %v", err)
		}
	}

	return db, scope
}

func seedVisibleProduct(t *testing.T, db *gorm.DB, id int64, verifyStatus, publishStatus int32) {
	t.Helper()

	for _, stmt := range []string{
		fmt.Sprintf(`INSERT INTO pms_product_spu(id, name, category_id, brand_id, main_pic, publish_status, verify_status, recommend_status, preview_status, stock, platform_id, tenant_id, merchant_id, is_deleted)
		 VALUES (%d, '测试商品', 11, 21, 'main.png', %d, %d, 0, 0, 8, 1, 10, 301, 0)`, id, publishStatus, verifyStatus),
		fmt.Sprintf(`INSERT INTO pms_product_sku(id, spu_id, stock, platform_id, tenant_id, merchant_id, is_deleted)
		 VALUES (%d, %d, 8, 1, 10, 301, 0)`, id+1000, id),
	} {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("seed product failed: %v", err)
		}
	}
}

func TestEnsureProductsReviewReadyRejectsDisabledMerchant(t *testing.T) {
	db, scope := newProductVisibilityTestDB(t)
	seedVisibleProduct(t, db, 1, ProductVerifyStatusPending, ProductPublishStatusOffShelf)

	if err := db.Exec(`UPDATE sys_merchant SET business_status = 2 WHERE id = 301`).Error; err != nil {
		t.Fatalf("disable merchant failed: %v", err)
	}

	err := EnsureProductsReviewReady(context.Background(), db, scope, []int64{1})
	if err == nil || !strings.Contains(err.Error(), "商户已停用") {
		t.Fatalf("expected merchant disabled error, got %v", err)
	}
}

func TestEnsureProductsPublishableRequiresApprovedStatus(t *testing.T) {
	db, scope := newProductVisibilityTestDB(t)
	seedVisibleProduct(t, db, 2, ProductVerifyStatusPending, ProductPublishStatusOffShelf)

	err := EnsureProductsPublishable(context.Background(), db, scope, []int64{2})
	if err == nil || !strings.Contains(err.Error(), "审核未通过") {
		t.Fatalf("expected approval error, got %v", err)
	}

	if err := db.Exec(`UPDATE pms_product_spu SET verify_status = 1 WHERE id = 2`).Error; err != nil {
		t.Fatalf("approve product failed: %v", err)
	}

	if err := EnsureProductsPublishable(context.Background(), db, scope, []int64{2}); err != nil {
		t.Fatalf("expected publishable product, got %v", err)
	}
}

func TestEnsureProductsOwnerActiveRejectsDisabledMerchant(t *testing.T) {
	db, scope := newProductVisibilityTestDB(t)
	seedVisibleProduct(t, db, 22, ProductVerifyStatusApproved, ProductPublishStatusOnShelf)

	if err := db.Exec(`UPDATE sys_merchant SET business_status = 2 WHERE id = 301`).Error; err != nil {
		t.Fatalf("disable merchant failed: %v", err)
	}

	err := EnsureProductsOwnerActive(context.Background(), db, scope, []int64{22})
	if err == nil || !strings.Contains(err.Error(), "商户已停用") {
		t.Fatalf("expected owner disabled error, got %v", err)
	}
}

func TestEnsureProductsRecommendableRequiresOnShelf(t *testing.T) {
	db, scope := newProductVisibilityTestDB(t)
	seedVisibleProduct(t, db, 3, ProductVerifyStatusApproved, ProductPublishStatusOffShelf)

	err := EnsureProductsRecommendable(context.Background(), db, scope, []int64{3})
	if err == nil || !strings.Contains(err.Error(), "未上架") {
		t.Fatalf("expected on-shelf error, got %v", err)
	}

	if err := db.Exec(`UPDATE pms_product_spu SET publish_status = 1 WHERE id = 3`).Error; err != nil {
		t.Fatalf("publish product failed: %v", err)
	}

	if err := EnsureProductsRecommendable(context.Background(), db, scope, []int64{3}); err != nil {
		t.Fatalf("expected recommendable product, got %v", err)
	}
}

func TestPartitionProductIndexIDsSeparatesVisibleAndInvisibleProducts(t *testing.T) {
	db, scope := newProductVisibilityTestDB(t)
	seedVisibleProduct(t, db, 11, ProductVerifyStatusApproved, ProductPublishStatusOnShelf)
	seedVisibleProduct(t, db, 12, ProductVerifyStatusApproved, ProductPublishStatusOffShelf)
	seedVisibleProduct(t, db, 13, ProductVerifyStatusRejected, ProductPublishStatusOnShelf)

	syncIDs, deleteIDs, err := PartitionProductIndexIDs(context.Background(), db, scope, []int64{11, 12, 13})
	if err != nil {
		t.Fatalf("partition index ids failed: %v", err)
	}
	if len(syncIDs) != 1 || syncIDs[0] != 11 {
		t.Fatalf("expected only on-shelf approved product to sync, got %+v", syncIDs)
	}
	if len(deleteIDs) != 2 || deleteIDs[0] != 12 || deleteIDs[1] != 13 {
		t.Fatalf("expected invisible products to delete, got %+v", deleteIDs)
	}
}

func TestPartitionProductIndexIDsTreatsDisabledMerchantAsInvisible(t *testing.T) {
	db, scope := newProductVisibilityTestDB(t)
	seedVisibleProduct(t, db, 21, ProductVerifyStatusApproved, ProductPublishStatusOnShelf)

	if err := db.Exec(`UPDATE sys_merchant SET business_status = 2 WHERE id = 301`).Error; err != nil {
		t.Fatalf("disable merchant failed: %v", err)
	}

	syncIDs, deleteIDs, err := PartitionProductIndexIDs(context.Background(), db, scope, []int64{21})
	if err != nil {
		t.Fatalf("partition index ids failed: %v", err)
	}
	if len(syncIDs) != 0 {
		t.Fatalf("expected no sync ids for disabled merchant, got %+v", syncIDs)
	}
	if len(deleteIDs) != 1 || deleteIDs[0] != 21 {
		t.Fatalf("expected product to be deleted from index, got %+v", deleteIDs)
	}
}
