package productspecservicelogic

import (
	"context"
	"strings"
	"testing"
	"time"

	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	"github.com/feihua/zero-admin/rpc/pms/gen/model"
	"github.com/feihua/zero-admin/rpc/pms/pmsclient"
)

func TestDeleteProductSpecRejectsSpecValueReference(t *testing.T) {
	svcCtx := newProductSpecScopeTestSvc(t)
	scope, _ := pkgscope.NormalizeGovernanceScope(pkgscope.SubjectTypeMerchant, 1, 10, 301)
	seedProductSpec(t, svcCtx.DB, model.PmsProductSpec{ID: 1, CategoryID: 100, Name: "颜色", Sort: 1, Status: 1, CreateBy: 1, CreateTime: time.Now(), IsDeleted: 0}, scope)
	if err := svcCtx.DB.Exec(`CREATE TABLE pms_product_spec_value (id INTEGER PRIMARY KEY, spec_id INTEGER NOT NULL, is_deleted INTEGER NOT NULL DEFAULT 0, platform_id INTEGER NOT NULL DEFAULT 1, tenant_id INTEGER NOT NULL DEFAULT 0, merchant_id INTEGER NOT NULL DEFAULT 0)`).Error; err != nil {
		t.Fatalf("create spec value table failed: %v", err)
	}
	if err := svcCtx.DB.Exec(`INSERT INTO pms_product_spec_value (id, spec_id, is_deleted, platform_id, tenant_id, merchant_id) VALUES (1, 1, 0, ?, ?, ?)`, scope.PlatformID, scope.TenantID, scope.MerchantID).Error; err != nil {
		t.Fatalf("seed spec value reference failed: %v", err)
	}

	logic := NewDeleteProductSpecLogic(context.Background(), svcCtx)
	_, err := logic.DeleteProductSpec(&pmsclient.DeleteProductSpecReq{Ids: []int64{1}, UpdateBy: 77, Scope: &pmsclient.GovernanceScope{ScopeType: scope.ScopeType, PlatformId: scope.PlatformID, TenantId: scope.TenantID, MerchantId: scope.MerchantID}})
	if err == nil || !strings.Contains(err.Error(), "商品规格已被规格值引用") {
		t.Fatalf("expected referenced spec error, got %v", err)
	}
}
