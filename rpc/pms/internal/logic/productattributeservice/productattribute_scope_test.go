package productattributeservicelogic

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

func newProductAttributeScopeTestSvc(t *testing.T) *svc.ServiceContext {
	dbPath := filepath.Join(t.TempDir(), "product-attribute-scope.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}
	if err := db.AutoMigrate(&model.PmsProductAttribute{}); err != nil {
		t.Fatalf("auto migrate attribute failed: %v", err)
	}
	for _, stmt := range []string{`ALTER TABLE pms_product_attribute ADD COLUMN platform_id INTEGER NOT NULL DEFAULT 1`, `ALTER TABLE pms_product_attribute ADD COLUMN tenant_id INTEGER NOT NULL DEFAULT 0`, `ALTER TABLE pms_product_attribute ADD COLUMN merchant_id INTEGER NOT NULL DEFAULT 0`} {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("alter attribute scope columns failed: %v", err)
		}
	}
	return &svc.ServiceContext{DB: db}
}

func seedProductAttribute(t *testing.T, db *gorm.DB, item model.PmsProductAttribute, current pkgscope.GovernanceScope) {
	if err := db.Create(&item).Error; err != nil {
		t.Fatalf("seed attribute failed: %v", err)
	}
	if err := db.Exec(`UPDATE pms_product_attribute SET platform_id=?, tenant_id=?, merchant_id=? WHERE id=?`, current.PlatformID, current.TenantID, current.MerchantID, item.ID).Error; err != nil {
		t.Fatalf("update attribute scope failed: %v", err)
	}
}

func TestQueryProductAttributeListFiltersByGovernanceScope(t *testing.T) {
	svcCtx := newProductAttributeScopeTestSvc(t)
	now := time.Now()
	merchantScope, _ := pkgscope.NormalizeGovernanceScope(pkgscope.SubjectTypeMerchant, 1, 10, 301)
	otherScope, _ := pkgscope.NormalizeGovernanceScope(pkgscope.SubjectTypeMerchant, 1, 10, 302)
	seedProductAttribute(t, svcCtx.DB, model.PmsProductAttribute{ID: 1, GroupID: 100, Name: "属性A", InputType: 1, ValueType: 1, InputList: "", Unit: "", IsRequired: 1, IsSearchable: 1, IsShow: 1, Sort: 1, Status: 1, CreateBy: 1, CreateTime: now, IsDeleted: 0}, merchantScope)
	seedProductAttribute(t, svcCtx.DB, model.PmsProductAttribute{ID: 2, GroupID: 100, Name: "属性B", InputType: 1, ValueType: 1, InputList: "", Unit: "", IsRequired: 1, IsSearchable: 1, IsShow: 1, Sort: 2, Status: 1, CreateBy: 1, CreateTime: now, IsDeleted: 0}, otherScope)
	logic := NewQueryProductAttributeListLogic(context.Background(), svcCtx)
	resp, err := logic.QueryProductAttributeList(&pmsclient.QueryProductAttributeListReq{PageNum: 1, PageSize: 10, Status: 2, IsRequired: 2, IsSearchable: 2, IsShow: 2, Scope: &pmsclient.GovernanceScope{ScopeType: merchantScope.ScopeType, PlatformId: merchantScope.PlatformID, TenantId: merchantScope.TenantID, MerchantId: merchantScope.MerchantID}})
	if err != nil {
		t.Fatalf("query attribute list failed: %v", err)
	}
	if resp.Total != 1 || len(resp.List) != 1 || resp.List[0].Id != 1 {
		t.Fatalf("unexpected scoped attribute list: %+v", resp)
	}
}

func TestQueryProductAttributeDetailRejectsCrossScopeLookup(t *testing.T) {
	svcCtx := newProductAttributeScopeTestSvc(t)
	now := time.Now()
	merchantScope, _ := pkgscope.NormalizeGovernanceScope(pkgscope.SubjectTypeMerchant, 1, 10, 301)
	otherScope, _ := pkgscope.NormalizeGovernanceScope(pkgscope.SubjectTypeMerchant, 1, 10, 302)
	seedProductAttribute(t, svcCtx.DB, model.PmsProductAttribute{ID: 11, GroupID: 100, Name: "越权属性", InputType: 1, ValueType: 1, InputList: "", Unit: "", IsRequired: 1, IsSearchable: 1, IsShow: 1, Sort: 1, Status: 1, CreateBy: 1, CreateTime: now, IsDeleted: 0}, merchantScope)
	logic := NewQueryProductAttributeDetailLogic(context.Background(), svcCtx)
	_, err := logic.QueryProductAttributeDetail(&pmsclient.QueryProductAttributeDetailReq{Id: 11, Scope: &pmsclient.GovernanceScope{ScopeType: otherScope.ScopeType, PlatformId: otherScope.PlatformID, TenantId: otherScope.TenantID, MerchantId: otherScope.MerchantID}})
	if err == nil || !strings.Contains(err.Error(), "商品属性不存在") {
		t.Fatalf("expected scoped not found, got %v", err)
	}
}
