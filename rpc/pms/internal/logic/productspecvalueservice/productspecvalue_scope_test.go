package productspecvalueservicelogic

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

func newProductSpecValueScopeTestSvc(t *testing.T) *svc.ServiceContext {
	dbPath := filepath.Join(t.TempDir(), "product-spec-value-scope.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}
	if err := db.AutoMigrate(&model.PmsProductSpecValue{}); err != nil {
		t.Fatalf("auto migrate spec value failed: %v", err)
	}
	for _, stmt := range []string{`ALTER TABLE pms_product_spec_value ADD COLUMN platform_id INTEGER NOT NULL DEFAULT 1`, `ALTER TABLE pms_product_spec_value ADD COLUMN tenant_id INTEGER NOT NULL DEFAULT 0`, `ALTER TABLE pms_product_spec_value ADD COLUMN merchant_id INTEGER NOT NULL DEFAULT 0`} {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("alter spec value scope columns failed: %v", err)
		}
	}
	return &svc.ServiceContext{DB: db}
}

func seedProductSpecValue(t *testing.T, db *gorm.DB, item model.PmsProductSpecValue, current pkgscope.GovernanceScope) {
	if err := db.Create(&item).Error; err != nil {
		t.Fatalf("seed spec value failed: %v", err)
	}
	if err := db.Exec(`UPDATE pms_product_spec_value SET platform_id=?, tenant_id=?, merchant_id=? WHERE id=?`, current.PlatformID, current.TenantID, current.MerchantID, item.ID).Error; err != nil {
		t.Fatalf("update spec value scope failed: %v", err)
	}
}

func TestQueryProductSpecValueListFiltersByGovernanceScope(t *testing.T) {
	svcCtx := newProductSpecValueScopeTestSvc(t)
	now := time.Now()
	merchantScope, _ := pkgscope.NormalizeGovernanceScope(pkgscope.SubjectTypeMerchant, 1, 10, 301)
	otherScope, _ := pkgscope.NormalizeGovernanceScope(pkgscope.SubjectTypeMerchant, 1, 10, 302)
	seedProductSpecValue(t, svcCtx.DB, model.PmsProductSpecValue{ID: 1, SpecID: 100, Value: "黑色", Sort: 1, Status: 1, CreateBy: 1, CreateTime: now, IsDeleted: 0}, merchantScope)
	seedProductSpecValue(t, svcCtx.DB, model.PmsProductSpecValue{ID: 2, SpecID: 100, Value: "白色", Sort: 2, Status: 1, CreateBy: 1, CreateTime: now, IsDeleted: 0}, otherScope)
	logic := NewQueryProductSpecValueListLogic(context.Background(), svcCtx)
	resp, err := logic.QueryProductSpecValueList(&pmsclient.QueryProductSpecValueListReq{PageNum: 1, PageSize: 10, Status: 2, Scope: &pmsclient.GovernanceScope{ScopeType: merchantScope.ScopeType, PlatformId: merchantScope.PlatformID, TenantId: merchantScope.TenantID, MerchantId: merchantScope.MerchantID}})
	if err != nil {
		t.Fatalf("query spec value list failed: %v", err)
	}
	if resp.Total != 1 || len(resp.List) != 1 || resp.List[0].Id != 1 {
		t.Fatalf("unexpected scoped spec value list: %+v", resp)
	}
}

func TestQueryProductSpecValueDetailRejectsCrossScopeLookup(t *testing.T) {
	svcCtx := newProductSpecValueScopeTestSvc(t)
	now := time.Now()
	merchantScope, _ := pkgscope.NormalizeGovernanceScope(pkgscope.SubjectTypeMerchant, 1, 10, 301)
	otherScope, _ := pkgscope.NormalizeGovernanceScope(pkgscope.SubjectTypeMerchant, 1, 10, 302)
	seedProductSpecValue(t, svcCtx.DB, model.PmsProductSpecValue{ID: 11, SpecID: 100, Value: "越权规格值", Sort: 1, Status: 1, CreateBy: 1, CreateTime: now, IsDeleted: 0}, merchantScope)
	logic := NewQueryProductSpecValueDetailLogic(context.Background(), svcCtx)
	_, err := logic.QueryProductSpecValueDetail(&pmsclient.QueryProductSpecValueDetailReq{Id: 11, Scope: &pmsclient.GovernanceScope{ScopeType: otherScope.ScopeType, PlatformId: otherScope.PlatformID, TenantId: otherScope.TenantID, MerchantId: otherScope.MerchantID}})
	if err == nil || !strings.Contains(err.Error(), "商品规格值不存在") {
		t.Fatalf("expected scoped not found, got %v", err)
	}
}
