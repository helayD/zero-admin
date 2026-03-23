package productattributegroupservicelogic

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

func newProductAttributeGroupScopeTestSvc(t *testing.T) *svc.ServiceContext {
	dbPath := filepath.Join(t.TempDir(), "product-attribute-group-scope.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}
	if err := db.AutoMigrate(&model.PmsProductAttributeGroup{}); err != nil {
		t.Fatalf("auto migrate attribute group failed: %v", err)
	}
	for _, stmt := range []string{`ALTER TABLE pms_product_attribute_group ADD COLUMN platform_id INTEGER NOT NULL DEFAULT 1`, `ALTER TABLE pms_product_attribute_group ADD COLUMN tenant_id INTEGER NOT NULL DEFAULT 0`, `ALTER TABLE pms_product_attribute_group ADD COLUMN merchant_id INTEGER NOT NULL DEFAULT 0`} {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("alter attribute group scope columns failed: %v", err)
		}
	}
	return &svc.ServiceContext{DB: db}
}

func seedProductAttributeGroup(t *testing.T, db *gorm.DB, item model.PmsProductAttributeGroup, current pkgscope.GovernanceScope) {
	if err := db.Create(&item).Error; err != nil {
		t.Fatalf("seed attribute group failed: %v", err)
	}
	if err := db.Exec(`UPDATE pms_product_attribute_group SET platform_id=?, tenant_id=?, merchant_id=? WHERE id=?`, current.PlatformID, current.TenantID, current.MerchantID, item.ID).Error; err != nil {
		t.Fatalf("update attribute group scope failed: %v", err)
	}
}

func TestQueryProductAttributeGroupListFiltersByGovernanceScope(t *testing.T) {
	svcCtx := newProductAttributeGroupScopeTestSvc(t)
	now := time.Now()
	merchantScope, _ := pkgscope.NormalizeGovernanceScope(pkgscope.SubjectTypeMerchant, 1, 10, 301)
	otherScope, _ := pkgscope.NormalizeGovernanceScope(pkgscope.SubjectTypeMerchant, 1, 10, 302)
	seedProductAttributeGroup(t, svcCtx.DB, model.PmsProductAttributeGroup{ID: 1, CategoryID: 100, Name: "分组A", Sort: 1, Status: 1, CreateBy: 1, CreateTime: now, IsDeleted: 0}, merchantScope)
	seedProductAttributeGroup(t, svcCtx.DB, model.PmsProductAttributeGroup{ID: 2, CategoryID: 100, Name: "分组B", Sort: 2, Status: 1, CreateBy: 1, CreateTime: now, IsDeleted: 0}, otherScope)
	logic := NewQueryProductAttributeGroupListLogic(context.Background(), svcCtx)
	resp, err := logic.QueryProductAttributeGroupList(&pmsclient.QueryProductAttributeGroupListReq{PageNum: 1, PageSize: 10, Status: 2, Scope: &pmsclient.GovernanceScope{ScopeType: merchantScope.ScopeType, PlatformId: merchantScope.PlatformID, TenantId: merchantScope.TenantID, MerchantId: merchantScope.MerchantID}})
	if err != nil {
		t.Fatalf("query attribute group list failed: %v", err)
	}
	if resp.Total != 1 || len(resp.List) != 1 || resp.List[0].Id != 1 {
		t.Fatalf("unexpected scoped attribute group list: %+v", resp)
	}
}

func TestQueryProductAttributeGroupDetailRejectsCrossScopeLookup(t *testing.T) {
	svcCtx := newProductAttributeGroupScopeTestSvc(t)
	now := time.Now()
	merchantScope, _ := pkgscope.NormalizeGovernanceScope(pkgscope.SubjectTypeMerchant, 1, 10, 301)
	otherScope, _ := pkgscope.NormalizeGovernanceScope(pkgscope.SubjectTypeMerchant, 1, 10, 302)
	seedProductAttributeGroup(t, svcCtx.DB, model.PmsProductAttributeGroup{ID: 11, CategoryID: 100, Name: "越权分组", Sort: 1, Status: 1, CreateBy: 1, CreateTime: now, IsDeleted: 0}, merchantScope)
	logic := NewQueryProductAttributeGroupDetailLogic(context.Background(), svcCtx)
	_, err := logic.QueryProductAttributeGroupDetail(&pmsclient.QueryProductAttributeGroupDetailReq{Id: 11, Scope: &pmsclient.GovernanceScope{ScopeType: otherScope.ScopeType, PlatformId: otherScope.PlatformID, TenantId: otherScope.TenantID, MerchantId: otherScope.MerchantID}})
	if err == nil || !strings.Contains(err.Error(), "商品属性分组不存在") {
		t.Fatalf("expected scoped not found, got %v", err)
	}
}
