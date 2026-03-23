package productcategoryservicelogic

import (
	"context"
	"strings"
	"testing"
	"time"

	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	"github.com/feihua/zero-admin/rpc/pms/gen/model"
	"github.com/feihua/zero-admin/rpc/pms/pmsclient"
)

func TestAddProductCategoryRejectsMissingAttributeBinding(t *testing.T) {
	svcCtx := newProductCategoryScopeTestSvc(t)
	if err := svcCtx.DB.Exec(`CREATE TABLE pms_product_attribute (id INTEGER PRIMARY KEY, status INTEGER NOT NULL DEFAULT 1, is_deleted INTEGER NOT NULL DEFAULT 0, platform_id INTEGER NOT NULL DEFAULT 1, tenant_id INTEGER NOT NULL DEFAULT 0, merchant_id INTEGER NOT NULL DEFAULT 0)`).Error; err != nil {
		t.Fatalf("create attribute table failed: %v", err)
	}
	if err := svcCtx.DB.Exec(`CREATE TABLE pms_product_category_attribute_relation (id INTEGER PRIMARY KEY AUTOINCREMENT, product_category_id INTEGER NOT NULL, product_attribute_id INTEGER NOT NULL, is_deleted INTEGER NOT NULL DEFAULT 0, platform_id INTEGER NOT NULL DEFAULT 1, tenant_id INTEGER NOT NULL DEFAULT 0, merchant_id INTEGER NOT NULL DEFAULT 0)`).Error; err != nil {
		t.Fatalf("create category attribute relation table failed: %v", err)
	}
	scope, _ := pkgscope.NormalizeGovernanceScope(pkgscope.SubjectTypeMerchant, 1, 10, 301)
	logic := NewAddProductCategoryLogic(context.Background(), svcCtx)
	_, err := logic.AddProductCategory(&pmsclient.AddProductCategoryReq{ParentId: 0, Name: "手机配件", Level: 0, ProductUnit: "件", NavStatus: 1, Sort: 1, Icon: "", Keywords: "", Description: "", IsEnabled: 1, ProductAttributeIdList: []int64{99}, CreateBy: 77, Scope: &pmsclient.GovernanceScope{ScopeType: scope.ScopeType, PlatformId: scope.PlatformID, TenantId: scope.TenantID, MerchantId: scope.MerchantID}})
	if err == nil || !strings.Contains(err.Error(), "绑定的商品属性不存在") {
		t.Fatalf("expected missing attribute error, got %v", err)
	}
}

func TestAddProductCategoryRejectsCrossScopeAttributeBinding(t *testing.T) {
	svcCtx := newProductCategoryScopeTestSvc(t)
	if err := svcCtx.DB.Exec(`CREATE TABLE pms_product_attribute (id INTEGER PRIMARY KEY, status INTEGER NOT NULL DEFAULT 1, is_deleted INTEGER NOT NULL DEFAULT 0, platform_id INTEGER NOT NULL DEFAULT 1, tenant_id INTEGER NOT NULL DEFAULT 0, merchant_id INTEGER NOT NULL DEFAULT 0)`).Error; err != nil {
		t.Fatalf("create attribute table failed: %v", err)
	}
	if err := svcCtx.DB.Exec(`CREATE TABLE pms_product_category_attribute_relation (id INTEGER PRIMARY KEY AUTOINCREMENT, product_category_id INTEGER NOT NULL, product_attribute_id INTEGER NOT NULL, is_deleted INTEGER NOT NULL DEFAULT 0, platform_id INTEGER NOT NULL DEFAULT 1, tenant_id INTEGER NOT NULL DEFAULT 0, merchant_id INTEGER NOT NULL DEFAULT 0)`).Error; err != nil {
		t.Fatalf("create category attribute relation table failed: %v", err)
	}
	if err := svcCtx.DB.Exec(`INSERT INTO pms_product_attribute (id, status, is_deleted, platform_id, tenant_id, merchant_id) VALUES (12, 1, 0, 1, 10, 302)`).Error; err != nil {
		t.Fatalf("seed cross-scope attribute failed: %v", err)
	}
	scope, _ := pkgscope.NormalizeGovernanceScope(pkgscope.SubjectTypeMerchant, 1, 10, 301)
	logic := NewAddProductCategoryLogic(context.Background(), svcCtx)
	_, err := logic.AddProductCategory(&pmsclient.AddProductCategoryReq{ParentId: 0, Name: "手机配件", Level: 0, ProductUnit: "件", NavStatus: 1, Sort: 1, Icon: "", Keywords: "", Description: "", IsEnabled: 1, ProductAttributeIdList: []int64{12}, CreateBy: 77, Scope: &pmsclient.GovernanceScope{ScopeType: scope.ScopeType, PlatformId: scope.PlatformID, TenantId: scope.TenantID, MerchantId: scope.MerchantID}})
	if err == nil || !strings.Contains(err.Error(), "作用域不一致") {
		t.Fatalf("expected cross-scope attribute error, got %v", err)
	}
}

func TestAddProductCategoryRejectsDisabledAttributeBinding(t *testing.T) {
	svcCtx := newProductCategoryScopeTestSvc(t)
	if err := svcCtx.DB.Exec(`CREATE TABLE pms_product_attribute (id INTEGER PRIMARY KEY, status INTEGER NOT NULL DEFAULT 1, is_deleted INTEGER NOT NULL DEFAULT 0, platform_id INTEGER NOT NULL DEFAULT 1, tenant_id INTEGER NOT NULL DEFAULT 0, merchant_id INTEGER NOT NULL DEFAULT 0)`).Error; err != nil {
		t.Fatalf("create attribute table failed: %v", err)
	}
	if err := svcCtx.DB.Exec(`CREATE TABLE pms_product_category_attribute_relation (id INTEGER PRIMARY KEY AUTOINCREMENT, product_category_id INTEGER NOT NULL, product_attribute_id INTEGER NOT NULL, is_deleted INTEGER NOT NULL DEFAULT 0, platform_id INTEGER NOT NULL DEFAULT 1, tenant_id INTEGER NOT NULL DEFAULT 0, merchant_id INTEGER NOT NULL DEFAULT 0)`).Error; err != nil {
		t.Fatalf("create category attribute relation table failed: %v", err)
	}
	if err := svcCtx.DB.Exec(`INSERT INTO pms_product_attribute (id, status, is_deleted, platform_id, tenant_id, merchant_id) VALUES (13, 0, 0, 1, 10, 301)`).Error; err != nil {
		t.Fatalf("seed disabled attribute failed: %v", err)
	}
	scope, _ := pkgscope.NormalizeGovernanceScope(pkgscope.SubjectTypeMerchant, 1, 10, 301)
	logic := NewAddProductCategoryLogic(context.Background(), svcCtx)
	_, err := logic.AddProductCategory(&pmsclient.AddProductCategoryReq{ParentId: 0, Name: "手机配件", Level: 0, ProductUnit: "件", NavStatus: 1, Sort: 1, Icon: "", Keywords: "", Description: "", IsEnabled: 1, ProductAttributeIdList: []int64{13}, CreateBy: 77, Scope: &pmsclient.GovernanceScope{ScopeType: scope.ScopeType, PlatformId: scope.PlatformID, TenantId: scope.TenantID, MerchantId: scope.MerchantID}})
	if err == nil || !strings.Contains(err.Error(), "只能绑定启用状态") {
		t.Fatalf("expected disabled attribute error, got %v", err)
	}
}

func TestDeleteProductCategoryRejectsReferencedAttributeGroup(t *testing.T) {
	svcCtx := newProductCategoryScopeTestSvc(t)
	scope, _ := pkgscope.NormalizeGovernanceScope(pkgscope.SubjectTypeMerchant, 1, 10, 301)
	seedProductCategory(t, svcCtx.DB, model.PmsProductCategory{ID: 1, ParentID: 0, Name: "手机", Level: 0, ProductCount: 0, ProductUnit: "件", NavStatus: 1, Sort: 1, Icon: "", Keywords: "", Description: "", IsEnabled: 1, CreateBy: 1, CreateTime: time.Now(), IsDeleted: 0}, scope)
	if err := svcCtx.DB.Exec(`CREATE TABLE pms_product_attribute_group (id INTEGER PRIMARY KEY, category_id INTEGER NOT NULL, is_deleted INTEGER NOT NULL DEFAULT 0, platform_id INTEGER NOT NULL DEFAULT 1, tenant_id INTEGER NOT NULL DEFAULT 0, merchant_id INTEGER NOT NULL DEFAULT 0)`).Error; err != nil {
		t.Fatalf("create attribute group table failed: %v", err)
	}
	if err := svcCtx.DB.Exec(`INSERT INTO pms_product_attribute_group (id, category_id, is_deleted, platform_id, tenant_id, merchant_id) VALUES (1, 1, 0, ?, ?, ?)`, scope.PlatformID, scope.TenantID, scope.MerchantID).Error; err != nil {
		t.Fatalf("seed attribute group reference failed: %v", err)
	}

	logic := NewDeleteProductCategoryLogic(context.Background(), svcCtx)
	_, err := logic.DeleteProductCategory(&pmsclient.DeleteProductCategoryReq{Ids: []int64{1}, UpdateBy: 77, Scope: &pmsclient.GovernanceScope{ScopeType: scope.ScopeType, PlatformId: scope.PlatformID, TenantId: scope.TenantID, MerchantId: scope.MerchantID}})
	if err == nil || !strings.Contains(err.Error(), "属性分组引用") {
		t.Fatalf("expected referenced category error, got %v", err)
	}
}
