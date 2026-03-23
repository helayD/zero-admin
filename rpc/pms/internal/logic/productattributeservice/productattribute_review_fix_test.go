package productattributeservicelogic

import (
	"context"
	"strings"
	"testing"
	"time"

	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	"github.com/feihua/zero-admin/rpc/pms/gen/model"
	"github.com/feihua/zero-admin/rpc/pms/pmsclient"
)

func TestDeleteProductAttributeRejectsCategoryRelationReference(t *testing.T) {
	svcCtx := newProductAttributeScopeTestSvc(t)
	scope, _ := pkgscope.NormalizeGovernanceScope(pkgscope.SubjectTypeMerchant, 1, 10, 301)
	seedProductAttribute(t, svcCtx.DB, model.PmsProductAttribute{ID: 1, GroupID: 100, Name: "颜色", InputType: 1, ValueType: 1, InputList: "", Unit: "", IsRequired: 1, IsSearchable: 1, IsShow: 1, Sort: 1, Status: 1, CreateBy: 1, CreateTime: time.Now(), IsDeleted: 0}, scope)
	if err := svcCtx.DB.Exec(`CREATE TABLE pms_product_category_attribute_relation (id INTEGER PRIMARY KEY AUTOINCREMENT, product_category_id INTEGER NOT NULL, product_attribute_id INTEGER NOT NULL, is_deleted INTEGER NOT NULL DEFAULT 0, platform_id INTEGER NOT NULL DEFAULT 1, tenant_id INTEGER NOT NULL DEFAULT 0, merchant_id INTEGER NOT NULL DEFAULT 0)`).Error; err != nil {
		t.Fatalf("create category relation table failed: %v", err)
	}
	if err := svcCtx.DB.Exec(`INSERT INTO pms_product_category_attribute_relation (product_category_id, product_attribute_id, is_deleted, platform_id, tenant_id, merchant_id) VALUES (1, 1, 0, ?, ?, ?)`, scope.PlatformID, scope.TenantID, scope.MerchantID).Error; err != nil {
		t.Fatalf("seed category relation failed: %v", err)
	}

	logic := NewDeleteProductAttributeLogic(context.Background(), svcCtx)
	_, err := logic.DeleteProductAttribute(&pmsclient.DeleteProductAttributeReq{Ids: []int64{1}, UpdateBy: 77, Scope: &pmsclient.GovernanceScope{ScopeType: scope.ScopeType, PlatformId: scope.PlatformID, TenantId: scope.TenantID, MerchantId: scope.MerchantID}})
	if err == nil || !strings.Contains(err.Error(), "商品分类绑定") {
		t.Fatalf("expected referenced attribute error, got %v", err)
	}
}
