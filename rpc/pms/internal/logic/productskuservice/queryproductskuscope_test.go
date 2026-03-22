package productskuservicelogic

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

func newProductSkuScopeTestSvc(t *testing.T) *svc.ServiceContext {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "product-sku-scope.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}
	if err := db.AutoMigrate(&model.PmsProductSku{}); err != nil {
		t.Fatalf("auto migrate sku failed: %v", err)
	}
	for _, stmt := range []string{
		`ALTER TABLE pms_product_sku ADD COLUMN platform_id INTEGER NOT NULL DEFAULT 1`,
		`ALTER TABLE pms_product_sku ADD COLUMN tenant_id INTEGER NOT NULL DEFAULT 0`,
		`ALTER TABLE pms_product_sku ADD COLUMN merchant_id INTEGER NOT NULL DEFAULT 0`,
	} {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("alter sku scope columns failed: %v", err)
		}
	}

	return &svc.ServiceContext{DB: db}
}

func seedProductSku(t *testing.T, db *gorm.DB, item model.PmsProductSku, current pkgscope.GovernanceScope) {
	t.Helper()

	if err := db.Create(&item).Error; err != nil {
		t.Fatalf("seed sku failed: %v", err)
	}
	if err := db.Exec(
		`UPDATE pms_product_sku SET platform_id = ?, tenant_id = ?, merchant_id = ? WHERE id = ?`,
		current.PlatformID,
		current.TenantID,
		current.MerchantID,
		item.ID,
	).Error; err != nil {
		t.Fatalf("update sku scope failed: %v", err)
	}
}

func TestQueryProductSkuListFiltersByGovernanceScope(t *testing.T) {
	svcCtx := newProductSkuScopeTestSvc(t)
	now := time.Now()

	merchantScope, err := pkgscope.NormalizeGovernanceScope(pkgscope.SubjectTypeMerchant, 1, 10, 301)
	if err != nil {
		t.Fatalf("normalize merchant scope failed: %v", err)
	}
	otherScope, err := pkgscope.NormalizeGovernanceScope(pkgscope.SubjectTypeMerchant, 1, 10, 302)
	if err != nil {
		t.Fatalf("normalize other scope failed: %v", err)
	}

	seedProductSku(t, svcCtx.DB, model.PmsProductSku{
		ID:                 1,
		SpuID:              1001,
		Name:               "商户一号 SKU",
		SkuCode:            "SKU-301",
		MainPic:            "a.png",
		AlbumPics:          "a.png",
		Price:              199,
		PromotionPrice:     189,
		PromotionStartTime: &now,
		PromotionEndTime:   &now,
		Stock:              9,
		LowStock:           1,
		SpecData:           "{}",
		Weight:             1,
		PublishStatus:      1,
		VerifyStatus:       1,
		Sort:               1,
		Sales:              5,
		CreateBy:           1,
		CreateTime:         now,
		IsDeleted:          0,
	}, merchantScope)
	seedProductSku(t, svcCtx.DB, model.PmsProductSku{
		ID:                 2,
		SpuID:              1002,
		Name:               "商户二号 SKU",
		SkuCode:            "SKU-302",
		MainPic:            "b.png",
		AlbumPics:          "b.png",
		Price:              299,
		PromotionPrice:     279,
		PromotionStartTime: &now,
		PromotionEndTime:   &now,
		Stock:              8,
		LowStock:           1,
		SpecData:           "{}",
		Weight:             1,
		PublishStatus:      1,
		VerifyStatus:       1,
		Sort:               1,
		Sales:              6,
		CreateBy:           1,
		CreateTime:         now,
		IsDeleted:          0,
	}, otherScope)

	logic := NewQueryProductSkuListLogic(context.Background(), svcCtx)
	resp, err := logic.QueryProductSkuList(&pmsclient.QueryProductSkuListReq{
		PageNum:       1,
		PageSize:      10,
		PublishStatus: 2,
		VerifyStatus:  2,
		Scope: &pmsclient.GovernanceScope{
			ScopeType:  merchantScope.ScopeType,
			PlatformId: merchantScope.PlatformID,
			TenantId:   merchantScope.TenantID,
			MerchantId: merchantScope.MerchantID,
		},
	})
	if err != nil {
		t.Fatalf("query sku list failed: %v", err)
	}
	if resp.Total != 1 || len(resp.List) != 1 {
		t.Fatalf("expected one scoped sku, got total=%d len=%d", resp.Total, len(resp.List))
	}
	if resp.List[0].Id != 1 {
		t.Fatalf("expected sku 1, got %+v", resp.List[0])
	}
}

func TestQueryProductSkuDetailRejectsCrossScopeLookup(t *testing.T) {
	svcCtx := newProductSkuScopeTestSvc(t)
	now := time.Now()

	merchantScope, err := pkgscope.NormalizeGovernanceScope(pkgscope.SubjectTypeMerchant, 1, 10, 301)
	if err != nil {
		t.Fatalf("normalize merchant scope failed: %v", err)
	}
	otherScope, err := pkgscope.NormalizeGovernanceScope(pkgscope.SubjectTypeMerchant, 1, 10, 302)
	if err != nil {
		t.Fatalf("normalize other scope failed: %v", err)
	}

	seedProductSku(t, svcCtx.DB, model.PmsProductSku{
		ID:                 11,
		SpuID:              2001,
		Name:               "越权校验 SKU",
		SkuCode:            "SKU-X",
		MainPic:            "x.png",
		AlbumPics:          "x.png",
		Price:              99,
		PromotionPrice:     88,
		PromotionStartTime: &now,
		PromotionEndTime:   &now,
		Stock:              7,
		LowStock:           1,
		SpecData:           "{}",
		Weight:             1,
		PublishStatus:      1,
		VerifyStatus:       1,
		Sort:               1,
		Sales:              1,
		CreateBy:           1,
		CreateTime:         now,
		IsDeleted:          0,
	}, merchantScope)

	logic := NewQueryProductSkuDetailLogic(context.Background(), svcCtx)
	_, err = logic.QueryProductSkuDetail(&pmsclient.QueryProductSkuDetailReq{
		Id: 11,
		Scope: &pmsclient.GovernanceScope{
			ScopeType:  otherScope.ScopeType,
			PlatformId: otherScope.PlatformID,
			TenantId:   otherScope.TenantID,
			MerchantId: otherScope.MerchantID,
		},
	})
	if err == nil {
		t.Fatal("expected cross-scope sku detail to fail")
	}
	if !strings.Contains(err.Error(), "商品SKU不存在") {
		t.Fatalf("expected not found error, got %v", err)
	}
}
