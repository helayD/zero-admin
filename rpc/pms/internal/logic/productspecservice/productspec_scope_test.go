package productspecservicelogic

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

func newProductSpecScopeTestSvc(t *testing.T) *svc.ServiceContext {
	dbPath := filepath.Join(t.TempDir(), "product-spec-scope.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}
	if err := db.AutoMigrate(&model.PmsProductSpec{}); err != nil {
		t.Fatalf("auto migrate spec failed: %v", err)
	}
	for _, stmt := range []string{`ALTER TABLE pms_product_spec ADD COLUMN platform_id INTEGER NOT NULL DEFAULT 1`, `ALTER TABLE pms_product_spec ADD COLUMN tenant_id INTEGER NOT NULL DEFAULT 0`, `ALTER TABLE pms_product_spec ADD COLUMN merchant_id INTEGER NOT NULL DEFAULT 0`} {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("alter spec scope columns failed: %v", err)
		}
	}
	return &svc.ServiceContext{DB: db}
}

func seedProductSpec(t *testing.T, db *gorm.DB, item model.PmsProductSpec, current pkgscope.GovernanceScope) {
	if err := db.Create(&item).Error; err != nil {
		t.Fatalf("seed spec failed: %v", err)
	}
	if err := db.Exec(`UPDATE pms_product_spec SET platform_id=?, tenant_id=?, merchant_id=? WHERE id=?`, current.PlatformID, current.TenantID, current.MerchantID, item.ID).Error; err != nil {
		t.Fatalf("update spec scope failed: %v", err)
	}
}

func TestQueryProductSpecListFiltersByGovernanceScope(t *testing.T) {
	svcCtx := newProductSpecScopeTestSvc(t)
	now := time.Now()
	merchantScope, _ := pkgscope.NormalizeGovernanceScope(pkgscope.SubjectTypeMerchant, 1, 10, 301)
	otherScope, _ := pkgscope.NormalizeGovernanceScope(pkgscope.SubjectTypeMerchant, 1, 10, 302)
	seedProductSpec(t, svcCtx.DB, model.PmsProductSpec{ID: 1, CategoryID: 100, Name: "规格A", Sort: 1, Status: 1, CreateBy: 1, CreateTime: now, IsDeleted: 0}, merchantScope)
	seedProductSpec(t, svcCtx.DB, model.PmsProductSpec{ID: 2, CategoryID: 100, Name: "规格B", Sort: 2, Status: 1, CreateBy: 1, CreateTime: now, IsDeleted: 0}, otherScope)
	logic := NewQueryProductSpecListLogic(context.Background(), svcCtx)
	resp, err := logic.QueryProductSpecList(&pmsclient.QueryProductSpecListReq{PageNum: 1, PageSize: 10, Status: 2, Scope: &pmsclient.GovernanceScope{ScopeType: merchantScope.ScopeType, PlatformId: merchantScope.PlatformID, TenantId: merchantScope.TenantID, MerchantId: merchantScope.MerchantID}})
	if err != nil {
		t.Fatalf("query spec list failed: %v", err)
	}
	if resp.Total != 1 || len(resp.List) != 1 || resp.List[0].Id != 1 {
		t.Fatalf("unexpected scoped spec list: %+v", resp)
	}
}

func TestQueryProductSpecDetailRejectsCrossScopeLookup(t *testing.T) {
	svcCtx := newProductSpecScopeTestSvc(t)
	now := time.Now()
	merchantScope, _ := pkgscope.NormalizeGovernanceScope(pkgscope.SubjectTypeMerchant, 1, 10, 301)
	otherScope, _ := pkgscope.NormalizeGovernanceScope(pkgscope.SubjectTypeMerchant, 1, 10, 302)
	seedProductSpec(t, svcCtx.DB, model.PmsProductSpec{ID: 11, CategoryID: 100, Name: "越权规格", Sort: 1, Status: 1, CreateBy: 1, CreateTime: now, IsDeleted: 0}, merchantScope)
	logic := NewQueryProductSpecDetailLogic(context.Background(), svcCtx)
	_, err := logic.QueryProductSpecDetail(&pmsclient.QueryProductSpecDetailReq{Id: 11, Scope: &pmsclient.GovernanceScope{ScopeType: otherScope.ScopeType, PlatformId: otherScope.PlatformID, TenantId: otherScope.TenantID, MerchantId: otherScope.MerchantID}})
	if err == nil || !strings.Contains(err.Error(), "商品规格不存在") {
		t.Fatalf("expected scoped not found, got %v", err)
	}
}
