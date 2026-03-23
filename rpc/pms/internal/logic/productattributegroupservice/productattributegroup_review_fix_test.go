package productattributegroupservicelogic

import (
	"context"
	"strings"
	"testing"
	"time"

	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	"github.com/feihua/zero-admin/rpc/pms/gen/model"
	"github.com/feihua/zero-admin/rpc/pms/pmsclient"
)

func TestDeleteProductAttributeGroupRejectsAttributeReference(t *testing.T) {
	svcCtx := newProductAttributeGroupScopeTestSvc(t)
	scope, _ := pkgscope.NormalizeGovernanceScope(pkgscope.SubjectTypeMerchant, 1, 10, 301)
	seedProductAttributeGroup(t, svcCtx.DB, model.PmsProductAttributeGroup{ID: 1, CategoryID: 100, Name: "分组A", Sort: 1, Status: 1, CreateBy: 1, CreateTime: time.Now(), IsDeleted: 0}, scope)
	if err := svcCtx.DB.Exec(`CREATE TABLE pms_product_attribute (id INTEGER PRIMARY KEY, group_id INTEGER NOT NULL, is_deleted INTEGER NOT NULL DEFAULT 0, platform_id INTEGER NOT NULL DEFAULT 1, tenant_id INTEGER NOT NULL DEFAULT 0, merchant_id INTEGER NOT NULL DEFAULT 0)`).Error; err != nil {
		t.Fatalf("create attribute table failed: %v", err)
	}
	if err := svcCtx.DB.Exec(`INSERT INTO pms_product_attribute (id, group_id, is_deleted, platform_id, tenant_id, merchant_id) VALUES (1, 1, 0, ?, ?, ?)`, scope.PlatformID, scope.TenantID, scope.MerchantID).Error; err != nil {
		t.Fatalf("seed attribute reference failed: %v", err)
	}

	logic := NewDeleteProductAttributeGroupLogic(context.Background(), svcCtx)
	_, err := logic.DeleteProductAttributeGroup(&pmsclient.DeleteProductAttributeGroupReq{Ids: []int64{1}, UpdateBy: 77, Scope: &pmsclient.GovernanceScope{ScopeType: scope.ScopeType, PlatformId: scope.PlatformID, TenantId: scope.TenantID, MerchantId: scope.MerchantID}})
	if err == nil || !strings.Contains(err.Error(), "商品属性分组已被商品属性引用") {
		t.Fatalf("expected referenced attribute group error, got %v", err)
	}
}
