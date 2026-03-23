package productbrandservicelogic

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

func newProductBrandScopeTestSvc(t *testing.T) *svc.ServiceContext {
	dbPath := filepath.Join(t.TempDir(), "product-brand-scope.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}
	if err := db.AutoMigrate(&model.PmsProductBrand{}); err != nil {
		t.Fatalf("auto migrate brand failed: %v", err)
	}
	for _, stmt := range []string{`ALTER TABLE pms_product_brand ADD COLUMN platform_id INTEGER NOT NULL DEFAULT 1`, `ALTER TABLE pms_product_brand ADD COLUMN tenant_id INTEGER NOT NULL DEFAULT 0`, `ALTER TABLE pms_product_brand ADD COLUMN merchant_id INTEGER NOT NULL DEFAULT 0`} {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("alter brand scope columns failed: %v", err)
		}
	}
	return &svc.ServiceContext{DB: db}
}

func seedProductBrand(t *testing.T, db *gorm.DB, item model.PmsProductBrand, current pkgscope.GovernanceScope) {
	if err := db.Create(&item).Error; err != nil {
		t.Fatalf("seed brand failed: %v", err)
	}
	if err := db.Exec(`UPDATE pms_product_brand SET platform_id=?, tenant_id=?, merchant_id=? WHERE id=?`, current.PlatformID, current.TenantID, current.MerchantID, item.ID).Error; err != nil {
		t.Fatalf("update brand scope failed: %v", err)
	}
}

func TestQueryProductBrandListFiltersByGovernanceScope(t *testing.T) {
	svcCtx := newProductBrandScopeTestSvc(t)
	now := time.Now()
	merchantScope, _ := pkgscope.NormalizeGovernanceScope(pkgscope.SubjectTypeMerchant, 1, 10, 301)
	otherScope, _ := pkgscope.NormalizeGovernanceScope(pkgscope.SubjectTypeMerchant, 1, 10, 302)
	seedProductBrand(t, svcCtx.DB, model.PmsProductBrand{ID: 1, Name: "品牌A", Logo: "a", BigPic: "a", Description: "a", FirstLetter: "A", Sort: 1, RecommendStatus: 1, ProductCount: 0, ProductCommentCount: 0, IsEnabled: 1, CreateBy: 1, CreateTime: now, IsDeleted: 0}, merchantScope)
	seedProductBrand(t, svcCtx.DB, model.PmsProductBrand{ID: 2, Name: "品牌B", Logo: "b", BigPic: "b", Description: "b", FirstLetter: "B", Sort: 2, RecommendStatus: 1, ProductCount: 0, ProductCommentCount: 0, IsEnabled: 1, CreateBy: 1, CreateTime: now, IsDeleted: 0}, otherScope)
	logic := NewQueryProductBrandListLogic(context.Background(), svcCtx)
	resp, err := logic.QueryProductBrandList(&pmsclient.QueryProductBrandListReq{PageNum: 1, PageSize: 10, RecommendStatus: 2, IsEnabled: 2, Scope: &pmsclient.GovernanceScope{ScopeType: merchantScope.ScopeType, PlatformId: merchantScope.PlatformID, TenantId: merchantScope.TenantID, MerchantId: merchantScope.MerchantID}})
	if err != nil {
		t.Fatalf("query brand list failed: %v", err)
	}
	if resp.Total != 1 || len(resp.List) != 1 || resp.List[0].Id != 1 {
		t.Fatalf("unexpected scoped brand list: %+v", resp)
	}
}

func TestQueryProductBrandDetailRejectsCrossScopeLookup(t *testing.T) {
	svcCtx := newProductBrandScopeTestSvc(t)
	now := time.Now()
	merchantScope, _ := pkgscope.NormalizeGovernanceScope(pkgscope.SubjectTypeMerchant, 1, 10, 301)
	otherScope, _ := pkgscope.NormalizeGovernanceScope(pkgscope.SubjectTypeMerchant, 1, 10, 302)
	seedProductBrand(t, svcCtx.DB, model.PmsProductBrand{ID: 11, Name: "越权品牌", Logo: "x", BigPic: "x", Description: "x", FirstLetter: "X", Sort: 1, RecommendStatus: 1, ProductCount: 0, ProductCommentCount: 0, IsEnabled: 1, CreateBy: 1, CreateTime: now, IsDeleted: 0}, merchantScope)
	logic := NewQueryProductBrandDetailLogic(context.Background(), svcCtx)
	_, err := logic.QueryProductBrandDetail(&pmsclient.QueryProductBrandDetailReq{Id: 11, Scope: &pmsclient.GovernanceScope{ScopeType: otherScope.ScopeType, PlatformId: otherScope.PlatformID, TenantId: otherScope.TenantID, MerchantId: otherScope.MerchantID}})
	if err == nil || !strings.Contains(err.Error(), "商品品牌不存在") {
		t.Fatalf("expected scoped not found, got %v", err)
	}
}
