package productskuservicelogic

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	"github.com/feihua/zero-admin/rpc/pms/gen/model"
	"github.com/feihua/zero-admin/rpc/pms/internal/svc"
	"github.com/feihua/zero-admin/rpc/pms/pmsclient"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newProductSkuMaintainTestSvc(t *testing.T) (*svc.ServiceContext, pkgscope.GovernanceScope) {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "product-sku-maintain.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}
	if err := db.AutoMigrate(&model.PmsProductSpu{}, &model.PmsProductSku{}); err != nil {
		t.Fatalf("auto migrate spu/sku failed: %v", err)
	}
	for _, stmt := range []string{
		`ALTER TABLE pms_product_spu ADD COLUMN platform_id INTEGER NOT NULL DEFAULT 1`,
		`ALTER TABLE pms_product_spu ADD COLUMN tenant_id INTEGER NOT NULL DEFAULT 0`,
		`ALTER TABLE pms_product_spu ADD COLUMN merchant_id INTEGER NOT NULL DEFAULT 0`,
		`ALTER TABLE pms_product_sku ADD COLUMN platform_id INTEGER NOT NULL DEFAULT 1`,
		`ALTER TABLE pms_product_sku ADD COLUMN tenant_id INTEGER NOT NULL DEFAULT 0`,
		`ALTER TABLE pms_product_sku ADD COLUMN merchant_id INTEGER NOT NULL DEFAULT 0`,
		`CREATE TABLE pms_product_attribute_value (id INTEGER PRIMARY KEY, spu_id INTEGER NOT NULL, platform_id INTEGER NOT NULL DEFAULT 1, tenant_id INTEGER NOT NULL DEFAULT 0, merchant_id INTEGER NOT NULL DEFAULT 0)`,
		`CREATE TABLE pms_member_price (id INTEGER PRIMARY KEY, product_id INTEGER NOT NULL, platform_id INTEGER NOT NULL DEFAULT 1, tenant_id INTEGER NOT NULL DEFAULT 0, merchant_id INTEGER NOT NULL DEFAULT 0)`,
		`CREATE TABLE pms_product_ladder (id INTEGER PRIMARY KEY, product_id INTEGER NOT NULL, platform_id INTEGER NOT NULL DEFAULT 1, tenant_id INTEGER NOT NULL DEFAULT 0, merchant_id INTEGER NOT NULL DEFAULT 0)`,
		`CREATE TABLE pms_product_full_reduction (id INTEGER PRIMARY KEY, product_id INTEGER NOT NULL, platform_id INTEGER NOT NULL DEFAULT 1, tenant_id INTEGER NOT NULL DEFAULT 0, merchant_id INTEGER NOT NULL DEFAULT 0)`,
	} {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("alter scope columns failed: %v", err)
		}
	}
	scope, err := pkgscope.NormalizeGovernanceScope(pkgscope.SubjectTypeMerchant, 1, 10, 301)
	if err != nil {
		t.Fatalf("normalize scope failed: %v", err)
	}
	return &svc.ServiceContext{DB: db}, scope
}

func seedScopedSpu(t *testing.T, db *gorm.DB, scope pkgscope.GovernanceScope, spu model.PmsProductSpu) {
	t.Helper()
	if err := db.Create(&spu).Error; err != nil {
		t.Fatalf("seed spu failed: %v", err)
	}
	if err := db.Exec(`UPDATE pms_product_spu SET platform_id=?, tenant_id=?, merchant_id=? WHERE id=?`, scope.PlatformID, scope.TenantID, scope.MerchantID, spu.ID).Error; err != nil {
		t.Fatalf("update spu scope failed: %v", err)
	}
}

func seedScopedSku(t *testing.T, db *gorm.DB, scope pkgscope.GovernanceScope, sku model.PmsProductSku) {
	t.Helper()
	if err := db.Create(&sku).Error; err != nil {
		t.Fatalf("seed sku failed: %v", err)
	}
	if err := db.Exec(`UPDATE pms_product_sku SET platform_id=?, tenant_id=?, merchant_id=? WHERE id=?`, scope.PlatformID, scope.TenantID, scope.MerchantID, sku.ID).Error; err != nil {
		t.Fatalf("update sku scope failed: %v", err)
	}
}

func TestAddProductSkuValidatesAgainstSiblingSkusAndRefreshesParentSummary(t *testing.T) {
	svcCtx, scope := newProductSkuMaintainTestSvc(t)
	now := time.Now()
	seedScopedSpu(t, svcCtx.DB, scope, model.PmsProductSpu{ID: 1001, Name: "SPU-1001", ProductSn: "SPU-1001", MainPic: "main.png", PriceRange: "10.00", Stock: 3, LowStock: 1, CreateBy: 1, CreateTime: now})
	seedScopedSku(t, svcCtx.DB, scope, model.PmsProductSku{ID: 2001, SpuID: 1001, Name: "黑-L", SkuCode: "SPU1001-颜色-黑色-尺码-L", Price: 10, Stock: 3, LowStock: 1, SpecData: `{"颜色":"黑色","尺码":"L"}`, CreateBy: 1, CreateTime: now})

	logic := NewAddProductSkuLogic(context.Background(), svcCtx)
	_, err := logic.AddProductSku(&pmsclient.AddProductSkuReq{
		SpuId:     1001,
		Name:      "黑-L-重复",
		Price:     12,
		Stock:     2,
		LowStock:  1,
		SpecData:  `{"颜色":"黑色","尺码":"L"}`,
		CreateBy:  1,
		Scope:     &pmsclient.GovernanceScope{ScopeType: scope.ScopeType, PlatformId: scope.PlatformID, TenantId: scope.TenantID, MerchantId: scope.MerchantID},
	})
	if err == nil || !strings.Contains(err.Error(), "规格组合重复") {
		t.Fatalf("expected sibling spec conflict, got %v", err)
	}

	_, err = logic.AddProductSku(&pmsclient.AddProductSkuReq{
		SpuId:     1001,
		Name:      "白-L",
		Price:     20,
		Stock:     5,
		LowStock:  2,
		SpecData:  `{"颜色":"白色","尺码":"L"}`,
		CreateBy:  1,
		Scope:     &pmsclient.GovernanceScope{ScopeType: scope.ScopeType, PlatformId: scope.PlatformID, TenantId: scope.TenantID, MerchantId: scope.MerchantID},
	})
	if err != nil {
		t.Fatalf("add sku failed: %v", err)
	}

	var spu model.PmsProductSpu
	if err := svcCtx.DB.First(&spu, 1001).Error; err != nil {
		t.Fatalf("load spu failed: %v", err)
	}
	if spu.PriceRange != "10.00-20.00" || spu.Stock != 8 || spu.LowStock != 3 {
		t.Fatalf("unexpected spu summary: priceRange=%s stock=%d lowStock=%d", spu.PriceRange, spu.Stock, spu.LowStock)
	}

	var added model.PmsProductSku
	if err := svcCtx.DB.Where("spu_id = ? AND name = ?", 1001, "白-L").First(&added).Error; err != nil {
		t.Fatalf("load added sku failed: %v", err)
	}
	if added.SkuCode == "" || !strings.HasPrefix(added.SkuCode, "SPU1001-") {
		t.Fatalf("expected deterministic generated sku code, got %s", added.SkuCode)
	}
}

func TestUpdateProductSkuAllowsKeepingOwnCodeAndRefreshesParentSummary(t *testing.T) {
	svcCtx, scope := newProductSkuMaintainTestSvc(t)
	now := time.Now()
	seedScopedSpu(t, svcCtx.DB, scope, model.PmsProductSpu{ID: 1002, Name: "SPU-1002", ProductSn: "SPU-1002", MainPic: "main.png", PriceRange: "10.00-20.00", Stock: 8, LowStock: 3, CreateBy: 1, CreateTime: now})
	seedScopedSku(t, svcCtx.DB, scope, model.PmsProductSku{ID: 2101, SpuID: 1002, Name: "黑-L", SkuCode: "KEEP-CODE", Price: 10, Stock: 3, LowStock: 1, SpecData: `{"颜色":"黑色","尺码":"L"}`, CreateBy: 1, CreateTime: now})
	seedScopedSku(t, svcCtx.DB, scope, model.PmsProductSku{ID: 2102, SpuID: 1002, Name: "白-L", SkuCode: "WHITE-CODE", Price: 20, Stock: 5, LowStock: 2, SpecData: `{"颜色":"白色","尺码":"L"}`, CreateBy: 1, CreateTime: now})

	logic := NewUpdateProductSkuLogic(context.Background(), svcCtx)
	_, err := logic.UpdateProductSku(&pmsclient.UpdateProductSkuReq{
		Data: []*pmsclient.UpdateProductSkuData{{
			Id:        2101,
			SpuId:     1002,
			Name:      "黑-L-改价",
			SkuCode:   "KEEP-CODE",
			Price:     30,
			Stock:     7,
			LowStock:  2,
			SpecData:  `{"颜色":"黑色","尺码":"XL"}`,
			UpdateBy:  1,
		}},
		Scope: &pmsclient.GovernanceScope{ScopeType: scope.ScopeType, PlatformId: scope.PlatformID, TenantId: scope.TenantID, MerchantId: scope.MerchantID},
	})
	if err != nil {
		t.Fatalf("update sku failed: %v", err)
	}

	var updated model.PmsProductSku
	if err := svcCtx.DB.First(&updated, 2101).Error; err != nil {
		t.Fatalf("load updated sku failed: %v", err)
	}
	if updated.SkuCode != "KEEP-CODE" {
		t.Fatalf("expected keep own code, got %s", updated.SkuCode)
	}
	if updated.Name != "黑-L-改价" || updated.Stock != 7 {
		t.Fatalf("unexpected updated sku: %+v", updated)
	}

	var spu model.PmsProductSpu
	if err := svcCtx.DB.First(&spu, 1002).Error; err != nil {
		t.Fatalf("load spu failed: %v", err)
	}
	if spu.PriceRange != "20.00-30.00" || spu.Stock != 12 || spu.LowStock != 4 {
		t.Fatalf("unexpected spu summary after update: priceRange=%s stock=%d lowStock=%d", spu.PriceRange, spu.Stock, spu.LowStock)
	}
}
