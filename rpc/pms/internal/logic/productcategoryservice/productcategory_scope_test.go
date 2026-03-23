package productcategoryservicelogic

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

func newProductCategoryScopeTestSvc(t *testing.T) *svc.ServiceContext {
	dbPath := filepath.Join(t.TempDir(), "product-category-scope.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}
	if err := db.AutoMigrate(&model.PmsProductCategory{}); err != nil {
		t.Fatalf("auto migrate category failed: %v", err)
	}
	for _, stmt := range []string{
		`ALTER TABLE pms_product_category ADD COLUMN platform_id INTEGER NOT NULL DEFAULT 1`,
		`ALTER TABLE pms_product_category ADD COLUMN tenant_id INTEGER NOT NULL DEFAULT 0`,
		`ALTER TABLE pms_product_category ADD COLUMN merchant_id INTEGER NOT NULL DEFAULT 0`,
		`ALTER TABLE pms_product_category ADD COLUMN logo TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE pms_product_category ADD COLUMN big_pic TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE pms_product_category ADD COLUMN first_letter TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE pms_product_category ADD COLUMN recommend_status INTEGER NOT NULL DEFAULT 0`,
		`ALTER TABLE pms_product_category ADD COLUMN product_comment_count INTEGER NOT NULL DEFAULT 0`,
	} {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("alter category scope columns failed: %v", err)
		}
	}
	return &svc.ServiceContext{DB: db}
}

func seedProductCategory(t *testing.T, db *gorm.DB, item model.PmsProductCategory, current pkgscope.GovernanceScope) {
	if err := db.Create(&item).Error; err != nil {
		t.Fatalf("seed category failed: %v", err)
	}
	if err := db.Exec(`UPDATE pms_product_category SET platform_id=?, tenant_id=?, merchant_id=? WHERE id=?`, current.PlatformID, current.TenantID, current.MerchantID, item.ID).Error; err != nil {
		t.Fatalf("update category scope failed: %v", err)
	}
}

func TestQueryProductCategoryListFiltersByGovernanceScope(t *testing.T) {
	svcCtx := newProductCategoryScopeTestSvc(t)
	now := time.Now()
	merchantScope, _ := pkgscope.NormalizeGovernanceScope(pkgscope.SubjectTypeMerchant, 1, 10, 301)
	otherScope, _ := pkgscope.NormalizeGovernanceScope(pkgscope.SubjectTypeMerchant, 1, 10, 302)
	seedProductCategory(t, svcCtx.DB, model.PmsProductCategory{ID: 1, ParentID: 0, Name: "分类A", Level: 0, ProductCount: 0, ProductUnit: "件", NavStatus: 1, Sort: 1, Icon: "", Keywords: "", Description: "", IsEnabled: 1, CreateBy: 1, CreateTime: now, IsDeleted: 0}, merchantScope)
	seedProductCategory(t, svcCtx.DB, model.PmsProductCategory{ID: 2, ParentID: 0, Name: "分类B", Level: 0, ProductCount: 0, ProductUnit: "件", NavStatus: 1, Sort: 2, Icon: "", Keywords: "", Description: "", IsEnabled: 1, CreateBy: 1, CreateTime: now, IsDeleted: 0}, otherScope)
	logic := NewQueryProductCategoryListLogic(context.Background(), svcCtx)
	resp, err := logic.QueryProductCategoryList(&pmsclient.QueryProductCategoryListReq{PageNum: 1, PageSize: 10, ParentId: 1000, NavStatus: 2, IsEnabled: 2, Scope: &pmsclient.GovernanceScope{ScopeType: merchantScope.ScopeType, PlatformId: merchantScope.PlatformID, TenantId: merchantScope.TenantID, MerchantId: merchantScope.MerchantID}})
	if err != nil {
		t.Fatalf("query category list failed: %v", err)
	}
	if resp.Total != 1 || len(resp.List) != 1 || resp.List[0].Id != 1 {
		t.Fatalf("unexpected scoped category list: %+v", resp)
	}
}

func TestQueryProductCategoryDetailRejectsCrossScopeLookup(t *testing.T) {
	svcCtx := newProductCategoryScopeTestSvc(t)
	now := time.Now()
	merchantScope, _ := pkgscope.NormalizeGovernanceScope(pkgscope.SubjectTypeMerchant, 1, 10, 301)
	otherScope, _ := pkgscope.NormalizeGovernanceScope(pkgscope.SubjectTypeMerchant, 1, 10, 302)
	seedProductCategory(t, svcCtx.DB, model.PmsProductCategory{ID: 11, ParentID: 0, Name: "越权分类", Level: 0, ProductCount: 0, ProductUnit: "件", NavStatus: 1, Sort: 1, Icon: "", Keywords: "", Description: "", IsEnabled: 1, CreateBy: 1, CreateTime: now, IsDeleted: 0}, merchantScope)
	logic := NewQueryProductCategoryDetailLogic(context.Background(), svcCtx)
	_, err := logic.QueryProductCategoryDetail(&pmsclient.QueryProductCategoryDetailReq{Id: 11, Scope: &pmsclient.GovernanceScope{ScopeType: otherScope.ScopeType, PlatformId: otherScope.PlatformID, TenantId: otherScope.TenantID, MerchantId: otherScope.MerchantID}})
	if err == nil || !strings.Contains(err.Error(), "产品分类不存在") {
		t.Fatalf("expected scoped not found, got %v", err)
	}
}
