package productbrandservicelogic

import (
	"context"
	"strings"
	"testing"
	"time"

	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	"github.com/feihua/zero-admin/rpc/pms/gen/model"
	"github.com/feihua/zero-admin/rpc/pms/pmsclient"
)

func TestDeleteProductBrandRejectsReferencedSpu(t *testing.T) {
	svcCtx := newProductBrandScopeTestSvc(t)
	scope, _ := pkgscope.NormalizeGovernanceScope(pkgscope.SubjectTypeMerchant, 1, 10, 301)
	seedProductBrand(t, svcCtx.DB, model.PmsProductBrand{ID: 1, Name: "品牌A", Logo: "a", BigPic: "a", Description: "a", FirstLetter: "A", Sort: 1, RecommendStatus: 1, ProductCount: 0, ProductCommentCount: 0, IsEnabled: 1, CreateBy: 1, CreateTime: time.Now(), IsDeleted: 0}, scope)
	if err := svcCtx.DB.Exec(`CREATE TABLE pms_product_spu (id INTEGER PRIMARY KEY, brand_id INTEGER NOT NULL, is_deleted INTEGER NOT NULL DEFAULT 0, platform_id INTEGER NOT NULL DEFAULT 1, tenant_id INTEGER NOT NULL DEFAULT 0, merchant_id INTEGER NOT NULL DEFAULT 0)`).Error; err != nil {
		t.Fatalf("create spu table failed: %v", err)
	}
	if err := svcCtx.DB.Exec(`INSERT INTO pms_product_spu (id, brand_id, is_deleted, platform_id, tenant_id, merchant_id) VALUES (1, ?, 0, ?, ?, ?)`, 1, scope.PlatformID, scope.TenantID, scope.MerchantID).Error; err != nil {
		t.Fatalf("seed spu reference failed: %v", err)
	}

	logic := NewDeleteProductBrandLogic(context.Background(), svcCtx)
	_, err := logic.DeleteProductBrand(&pmsclient.DeleteProductBrandReq{Ids: []int64{1}, UpdateBy: 77, Scope: &pmsclient.GovernanceScope{ScopeType: scope.ScopeType, PlatformId: scope.PlatformID, TenantId: scope.TenantID, MerchantId: scope.MerchantID}})
	if err == nil || !strings.Contains(err.Error(), "商品品牌已被商品建档引用") {
		t.Fatalf("expected referenced brand error, got %v", err)
	}
}
