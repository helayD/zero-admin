package common

import (
	"context"
	"strings"
	"testing"

	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	"github.com/feihua/zero-admin/rpc/pms/pmsclient"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupDraftValidationDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}
	stmts := []string{
		`CREATE TABLE pms_product_category (id INTEGER PRIMARY KEY, is_deleted INTEGER DEFAULT 0, platform_id INTEGER, tenant_id INTEGER, merchant_id INTEGER)`,
		`CREATE TABLE pms_product_brand (id INTEGER PRIMARY KEY, is_deleted INTEGER DEFAULT 0, is_enabled INTEGER DEFAULT 1, platform_id INTEGER, tenant_id INTEGER, merchant_id INTEGER)`,
		`CREATE TABLE pms_product_attribute (id INTEGER PRIMARY KEY, is_deleted INTEGER DEFAULT 0, status INTEGER DEFAULT 1, platform_id INTEGER, tenant_id INTEGER, merchant_id INTEGER)`,
		`CREATE TABLE pms_product_sku (id INTEGER PRIMARY KEY AUTOINCREMENT, spu_id INTEGER, sku_code TEXT, is_deleted INTEGER DEFAULT 0, platform_id INTEGER, tenant_id INTEGER, merchant_id INTEGER)`,
	}
	for _, stmt := range stmts {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("exec stmt failed: %v", err)
		}
	}
	if err := db.Exec(`INSERT INTO pms_product_category(id, is_deleted, platform_id, tenant_id, merchant_id) VALUES (11,0,1,10,20), (12,0,1,99,88)`).Error; err != nil {
		t.Fatalf("seed category failed: %v", err)
	}
	if err := db.Exec(`INSERT INTO pms_product_brand(id, is_deleted, is_enabled, platform_id, tenant_id, merchant_id) VALUES (21,0,1,1,10,20), (22,0,0,1,10,20)`).Error; err != nil {
		t.Fatalf("seed brand failed: %v", err)
	}
	if err := db.Exec(`INSERT INTO pms_product_attribute(id, is_deleted, status, platform_id, tenant_id, merchant_id) VALUES (31,0,1,1,10,20), (32,0,0,1,10,20), (33,0,1,1,99,88)`).Error; err != nil {
		t.Fatalf("seed attribute failed: %v", err)
	}
	if err := db.Exec(`INSERT INTO pms_product_sku(id, spu_id, sku_code, is_deleted, platform_id, tenant_id, merchant_id) VALUES (1, 99, 'DUP-CODE', 0, 1,10,20)`).Error; err != nil {
		t.Fatalf("seed sku failed: %v", err)
	}
	return db
}

func TestValidateProductDraftSuccess(t *testing.T) {
	db := setupDraftValidationDB(t)
	scope := pkgscope.DefaultScope(1, 10, 20)
	resp, err := ValidateProductDraft(context.Background(), db, scope, &pmsclient.ProductSpuReq{
		Id:                        100,
		Name:                      "示例商品",
		ProductSn:                 "SPU-100",
		CategoryId:                11,
		BrandId:                   21,
		MainPic:                   "https://img.example.com/main.png",
		FulfillmentMode:           "physical_delivery",
		ProductAttributeValueList: []*pmsclient.ProductAttributeValueList{{ProductAttributeId: 31, AttributeValues: "黑色"}},
		SkuStockList:              []*pmsclient.SkuStockList{{Name: "黑色-L", Price: 199, PromotionPrice: 99, Stock: 10, LowStock: 2, SpecData: `{"颜色":"黑色","尺码":"L"}`}, {Name: "白色-L", Price: 299, Stock: 20, LowStock: 4, SpecData: `{"颜色":"白色","尺码":"L"}`}},
	})
	if err != nil {
		t.Fatalf("ValidateProductDraft failed: %v", err)
	}
	if resp.PriceRange != "199.00-299.00" {
		t.Fatalf("unexpected price range: %s", resp.PriceRange)
	}
	if resp.TotalStock != 30 || resp.LowStock != 6 {
		t.Fatalf("unexpected stock summary: %+v", resp)
	}
}

func TestValidateProductDraftRejectsInvalidCases(t *testing.T) {
	db := setupDraftValidationDB(t)
	scope := pkgscope.DefaultScope(1, 10, 20)
	cases := []struct {
		name string
		req  *pmsclient.ProductSpuReq
		want string
	}{
		{
			name: "empty fulfillment mode rejected",
			req:  &pmsclient.ProductSpuReq{Id: 1, Name: "商品", ProductSn: "SPU-1", CategoryId: 11, BrandId: 21, MainPic: "a", SkuStockList: []*pmsclient.SkuStockList{{Name: "A", Price: 10, Stock: 1, SpecData: `{"颜色":"黑"}`}}},
			want: "履约模式不能为空",
		},
		{
			name: "digital_asset without rule_id rejected",
			req:  &pmsclient.ProductSpuReq{Id: 1, Name: "商品", ProductSn: "SPU-1", CategoryId: 11, BrandId: 21, MainPic: "a", FulfillmentMode: "digital_asset", SkuStockList: []*pmsclient.SkuStockList{{Name: "A", Price: 10, Stock: 1, SpecData: `{"颜色":"黑"}`}}},
			want: "提货卡模式商品必须绑定发卡规则",
		},
		{
			name: "duplicate spec",
			req:  &pmsclient.ProductSpuReq{Id: 1, Name: "商品", ProductSn: "SPU-1", CategoryId: 11, BrandId: 21, MainPic: "a", FulfillmentMode: "physical_delivery", SkuStockList: []*pmsclient.SkuStockList{{Name: "A", Price: 10, Stock: 1, SpecData: `{"颜色":"黑"}`}, {Name: "B", Price: 12, Stock: 1, SpecData: `{"颜色":"黑"}`}}},
			want: "规格组合重复",
		},
		{
			name: "illegal stock",
			req:  &pmsclient.ProductSpuReq{Id: 1, Name: "商品", ProductSn: "SPU-1", CategoryId: 11, BrandId: 21, MainPic: "a", FulfillmentMode: "physical_delivery", SkuStockList: []*pmsclient.SkuStockList{{Name: "A", Price: 10, Stock: 0, LowStock: 1, SpecData: `{"颜色":"黑"}`}}},
			want: "预警库存不能大于可用库存",
		},
		{
			name: "disabled brand",
			req:  &pmsclient.ProductSpuReq{Id: 1, Name: "商品", ProductSn: "SPU-1", CategoryId: 11, BrandId: 22, MainPic: "a", FulfillmentMode: "physical_delivery", SkuStockList: []*pmsclient.SkuStockList{{Name: "A", Price: 10, Stock: 1, SpecData: `{"颜色":"黑"}`}}},
			want: "商品品牌已失效",
		},
		{
			name: "cross scope category",
			req:  &pmsclient.ProductSpuReq{Id: 1, Name: "商品", ProductSn: "SPU-1", CategoryId: 12, BrandId: 21, MainPic: "a", FulfillmentMode: "physical_delivery", SkuStockList: []*pmsclient.SkuStockList{{Name: "A", Price: 10, Stock: 1, SpecData: `{"颜色":"黑"}`}}},
			want: "当前主体无权将商品归属到该商品分类",
		},
		{
			name: "cross scope attribute",
			req: &pmsclient.ProductSpuReq{Id: 1, Name: "商品", ProductSn: "SPU-1", CategoryId: 11, BrandId: 21, MainPic: "a", FulfillmentMode: "physical_delivery",
				ProductAttributeValueList: []*pmsclient.ProductAttributeValueList{{ProductAttributeId: 33, AttributeValues: "越权属性"}},
				SkuStockList:              []*pmsclient.SkuStockList{{Name: "A", Price: 10, Stock: 1, SpecData: `{"颜色":"黑"}`}},
			},
			want: "当前主体无权绑定该商品属性",
		},
		{
			name: "empty stock on all sku",
			req:  &pmsclient.ProductSpuReq{Id: 1, Name: "商品", ProductSn: "SPU-1", CategoryId: 11, BrandId: 21, MainPic: "a", FulfillmentMode: "physical_delivery", SkuStockList: []*pmsclient.SkuStockList{{Name: "A", Price: 10, Stock: 0, LowStock: 0, SpecData: `{"颜色":"黑"}`}}},
			want: "至少一个SKU需要具备有效库存",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ValidateProductDraft(context.Background(), db, scope, tc.req)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("expected %q, got %v", tc.want, err)
			}
		})
	}
}

func TestValidateSkuDraftsRejectsSkuCodeConflict(t *testing.T) {
	db := setupDraftValidationDB(t)
	scope := pkgscope.DefaultScope(1, 10, 20)
	_, err := ValidateSkuDrafts(context.Background(), db, scope, []ProductDraftSku{{Name: "SKU-A", SkuCode: "DUP-CODE", Price: 88, Stock: 2, SpecData: `{"颜色":"黑"}`}}, 0)
	if err == nil || !strings.Contains(err.Error(), "SKU编码冲突") {
		t.Fatalf("expected sku code conflict, got %v", err)
	}
}

func TestValidateSkuDraftsAllowsCurrentSkuAndWholeSpuReplacement(t *testing.T) {
	db := setupDraftValidationDB(t)
	scope := pkgscope.DefaultScope(1, 10, 20)

	_, err := ValidateSkuDrafts(context.Background(), db, scope, []ProductDraftSku{{ID: 1, Name: "SKU-A", SkuCode: "DUP-CODE", Price: 88, Stock: 2, SpecData: `{"颜色":"黑"}`}}, 0)
	if err != nil {
		t.Fatalf("expected current sku update to pass, got %v", err)
	}

	_, err = ValidateSkuDrafts(context.Background(), db, scope, []ProductDraftSku{{Name: "SKU-A", SkuCode: "DUP-CODE", Price: 88, Stock: 2, SpecData: `{"颜色":"黑"}`}}, 99)
	if err != nil {
		t.Fatalf("expected whole spu replacement to pass, got %v", err)
	}
}
