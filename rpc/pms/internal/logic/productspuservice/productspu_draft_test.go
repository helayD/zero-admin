package productspuservicelogic

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	"github.com/feihua/zero-admin/rpc/pms/gen/model"
	"github.com/feihua/zero-admin/rpc/pms/gen/query"
	"github.com/feihua/zero-admin/rpc/pms/internal/svc"
	"github.com/feihua/zero-admin/rpc/pms/pmsclient"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type mockProductVertifyRecordModel struct {
	inserted []*model.ProductVertifyRecord
}

func (m *mockProductVertifyRecordModel) Insert(ctx context.Context, data *model.ProductVertifyRecord) error {
	copied := *data
	m.inserted = append(m.inserted, &copied)
	return nil
}

func (m *mockProductVertifyRecordModel) FindOne(ctx context.Context, id string) (*model.ProductVertifyRecord, error) {
	if len(m.inserted) == 0 {
		return nil, model.ErrNotFound
	}
	return m.inserted[0], nil
}

func (m *mockProductVertifyRecordModel) Update(ctx context.Context, data *model.ProductVertifyRecord) (*mongo.UpdateResult, error) {
	return &mongo.UpdateResult{}, nil
}

func (m *mockProductVertifyRecordModel) Delete(ctx context.Context, id string) (int64, error) {
	return 0, nil
}

func (m *mockProductVertifyRecordModel) FindAll(ctx context.Context, productId int64) ([]*model.ProductVertifyRecord, error) {
	return m.inserted, nil
}

func newProductSpuDraftTestSvc(t *testing.T) (*svc.ServiceContext, pkgscope.GovernanceScope) {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "product-spu-draft.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}
	query.SetDefault(db)
	if err := db.AutoMigrate(
		&model.PmsProductSpu{},
		&model.PmsProductSku{},
		&model.PmsMemberPrice{},
		&model.PmsProductLadder{},
		&model.PmsProductFullReduction{},
		&model.PmsProductAttributeValue{},
	); err != nil {
		t.Fatalf("auto migrate spu tables failed: %v", err)
	}
	for _, stmt := range []string{
		`ALTER TABLE pms_product_spu ADD COLUMN platform_id INTEGER NOT NULL DEFAULT 1`,
		`ALTER TABLE pms_product_spu ADD COLUMN tenant_id INTEGER NOT NULL DEFAULT 0`,
		`ALTER TABLE pms_product_spu ADD COLUMN merchant_id INTEGER NOT NULL DEFAULT 0`,
		`ALTER TABLE pms_product_sku ADD COLUMN platform_id INTEGER NOT NULL DEFAULT 1`,
		`ALTER TABLE pms_product_sku ADD COLUMN tenant_id INTEGER NOT NULL DEFAULT 0`,
		`ALTER TABLE pms_product_sku ADD COLUMN merchant_id INTEGER NOT NULL DEFAULT 0`,
		`ALTER TABLE pms_member_price ADD COLUMN platform_id INTEGER NOT NULL DEFAULT 1`,
		`ALTER TABLE pms_member_price ADD COLUMN tenant_id INTEGER NOT NULL DEFAULT 0`,
		`ALTER TABLE pms_member_price ADD COLUMN merchant_id INTEGER NOT NULL DEFAULT 0`,
		`ALTER TABLE pms_product_ladder ADD COLUMN platform_id INTEGER NOT NULL DEFAULT 1`,
		`ALTER TABLE pms_product_ladder ADD COLUMN tenant_id INTEGER NOT NULL DEFAULT 0`,
		`ALTER TABLE pms_product_ladder ADD COLUMN merchant_id INTEGER NOT NULL DEFAULT 0`,
		`ALTER TABLE pms_product_full_reduction ADD COLUMN platform_id INTEGER NOT NULL DEFAULT 1`,
		`ALTER TABLE pms_product_full_reduction ADD COLUMN tenant_id INTEGER NOT NULL DEFAULT 0`,
		`ALTER TABLE pms_product_full_reduction ADD COLUMN merchant_id INTEGER NOT NULL DEFAULT 0`,
		`ALTER TABLE pms_product_attribute_value ADD COLUMN platform_id INTEGER NOT NULL DEFAULT 1`,
		`ALTER TABLE pms_product_attribute_value ADD COLUMN tenant_id INTEGER NOT NULL DEFAULT 0`,
		`ALTER TABLE pms_product_attribute_value ADD COLUMN merchant_id INTEGER NOT NULL DEFAULT 0`,
	} {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("alter scope columns failed: %v", err)
		}
	}
	for _, stmt := range []string{
		`CREATE TABLE pms_product_category (id INTEGER PRIMARY KEY, is_enabled INTEGER NOT NULL DEFAULT 1, is_deleted INTEGER NOT NULL DEFAULT 0, platform_id INTEGER NOT NULL DEFAULT 1, tenant_id INTEGER NOT NULL DEFAULT 0, merchant_id INTEGER NOT NULL DEFAULT 0)`,
		`CREATE TABLE pms_product_brand (id INTEGER PRIMARY KEY, is_enabled INTEGER NOT NULL DEFAULT 1, is_deleted INTEGER NOT NULL DEFAULT 0, platform_id INTEGER NOT NULL DEFAULT 1, tenant_id INTEGER NOT NULL DEFAULT 0, merchant_id INTEGER NOT NULL DEFAULT 0)`,
		`CREATE TABLE pms_product_attribute (id INTEGER PRIMARY KEY, status INTEGER NOT NULL DEFAULT 1, is_deleted INTEGER NOT NULL DEFAULT 0, platform_id INTEGER NOT NULL DEFAULT 1, tenant_id INTEGER NOT NULL DEFAULT 0, merchant_id INTEGER NOT NULL DEFAULT 0)`,
	} {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("exec schema failed: %v", err)
		}
	}
	scope, err := pkgscope.NormalizeGovernanceScope(pkgscope.SubjectTypeMerchant, 1, 10, 301)
	if err != nil {
		t.Fatalf("normalize scope failed: %v", err)
	}
	for _, stmt := range []string{
		`INSERT INTO pms_product_category(id, is_enabled, is_deleted, platform_id, tenant_id, merchant_id) VALUES (11, 1, 0, 1, 10, 301)`,
		`INSERT INTO pms_product_brand(id, is_enabled, is_deleted, platform_id, tenant_id, merchant_id) VALUES (21, 1, 0, 1, 10, 301)`,
		`INSERT INTO pms_product_attribute(id, status, is_deleted, platform_id, tenant_id, merchant_id) VALUES (31, 1, 0, 1, 10, 301)`,
	} {
		if err := db.Exec(stmt).Error; err != nil {
			t.Fatalf("seed catalog failed: %v", err)
		}
	}

	return &svc.ServiceContext{DB: db}, scope
}

func TestAddProductSpuGeneratesDeterministicSkuCodesAndScopeSummary(t *testing.T) {
	svcCtx, scope := newProductSpuDraftTestSvc(t)
	logic := NewAddProductSpuLogic(context.Background(), svcCtx)

	resp, err := logic.AddProductSpu(&pmsclient.ProductSpuReq{
		Name:         "测试商品",
		ProductSn:    "SPU-NEW-1",
		CategoryId:   11,
		CategoryName: "测试分类",
		BrandId:      21,
		BrandName:    "测试品牌",
		Unit:         "件",
		MainPic:      "main.png",
		AlbumPics:    "a.png,b.png",
		CreateBy:     1001,
		Scope:        &pmsclient.GovernanceScope{ScopeType: scope.ScopeType, PlatformId: scope.PlatformID, TenantId: scope.TenantID, MerchantId: scope.MerchantID},
		MemberPriceList: []*pmsclient.MemberPriceList{{LevelId: 1, Price: 88, LevelName: "黄金会员"}},
		ProductLadderList: []*pmsclient.ProductLadderList{{Count: 2, Discount: 90, Price: 89}},
		ProductFullReductionList: []*pmsclient.ProductFullReductionList{{FullPrice: 199, ReducePrice: 20}},
		ProductAttributeValueList: []*pmsclient.ProductAttributeValueList{{ProductAttributeId: 31, AttributeValues: "黑色"}},
		SkuStockList: []*pmsclient.SkuStockList{
			{Name: "黑色-L", Price: 99, PromotionPrice: 79, Stock: 8, LowStock: 2, SpecData: `{"颜色":"黑色","尺码":"L"}`},
			{Name: "白色-L", Price: 129, Stock: 5, LowStock: 1, SpecData: `{"颜色":"白色","尺码":"L"}`},
		},
	})
	if err != nil {
		t.Fatalf("add product spu failed: %v", err)
	}
	if resp.SpuId <= 0 {
		t.Fatalf("expected positive spu id, got %+v", resp)
	}

	var spu model.PmsProductSpu
	if err := svcCtx.DB.First(&spu, resp.SpuId).Error; err != nil {
		t.Fatalf("load spu failed: %v", err)
	}
	if spu.PriceRange != "99.00-129.00" || spu.Stock != 13 || spu.LowStock != 3 {
		t.Fatalf("unexpected spu summary: priceRange=%s stock=%d lowStock=%d", spu.PriceRange, spu.Stock, spu.LowStock)
	}
	var spuScope struct {
		PlatformID int64 `gorm:"column:platform_id"`
		TenantID   int64 `gorm:"column:tenant_id"`
		MerchantID int64 `gorm:"column:merchant_id"`
	}
	if err := svcCtx.DB.Table("pms_product_spu").Select("platform_id, tenant_id, merchant_id").Where("id = ?", resp.SpuId).Take(&spuScope).Error; err != nil {
		t.Fatalf("load spu scope failed: %v", err)
	}
	if spuScope.PlatformID != scope.PlatformID || spuScope.TenantID != scope.TenantID || spuScope.MerchantID != scope.MerchantID {
		t.Fatalf("unexpected spu scope: %+v", spuScope)
	}

	var skus []model.PmsProductSku
	if err := svcCtx.DB.Where("spu_id = ?", resp.SpuId).Order("id asc").Find(&skus).Error; err != nil {
		t.Fatalf("load skus failed: %v", err)
	}
	if len(skus) != 2 {
		t.Fatalf("expected 2 skus, got %d", len(skus))
	}
	for _, sku := range skus {
		if !strings.HasPrefix(sku.SkuCode, "SPU") {
			t.Fatalf("expected generated sku code, got %s", sku.SkuCode)
		}
		var skuScope struct {
			PlatformID int64 `gorm:"column:platform_id"`
			TenantID   int64 `gorm:"column:tenant_id"`
			MerchantID int64 `gorm:"column:merchant_id"`
		}
		if err := svcCtx.DB.Table("pms_product_sku").Select("platform_id, tenant_id, merchant_id").Where("id = ?", sku.ID).Take(&skuScope).Error; err != nil {
			t.Fatalf("load sku scope failed: %v", err)
		}
		if skuScope.PlatformID != scope.PlatformID || skuScope.TenantID != scope.TenantID || skuScope.MerchantID != scope.MerchantID {
			t.Fatalf("unexpected sku scope: %+v", skuScope)
		}
	}

	for _, target := range []struct {
		table  string
		column string
		label  string
	}{
		{table: "pms_member_price", column: "product_id", label: "member price"},
		{table: "pms_product_ladder", column: "product_id", label: "product ladder"},
		{table: "pms_product_full_reduction", column: "product_id", label: "full reduction"},
		{table: "pms_product_attribute_value", column: "spu_id", label: "attribute value"},
	} {
		var rows []struct {
			PlatformID int64 `gorm:"column:platform_id"`
			TenantID   int64 `gorm:"column:tenant_id"`
			MerchantID int64 `gorm:"column:merchant_id"`
		}
		if err := svcCtx.DB.Table(target.table).Select("platform_id, tenant_id, merchant_id").Where(target.column+" = ?", resp.SpuId).Find(&rows).Error; err != nil {
			t.Fatalf("load %s scope failed: %v", target.label, err)
		}
		for _, row := range rows {
			if row.PlatformID != scope.PlatformID || row.TenantID != scope.TenantID || row.MerchantID != scope.MerchantID {
				t.Fatalf("unexpected %s scope: %+v", target.label, row)
			}
		}
	}
}

func TestUpdateProductSpuRebuildsSkuSummaryWithDeterministicCodes(t *testing.T) {
	svcCtx, scope := newProductSpuDraftTestSvc(t)
	now := time.Now()
	if err := svcCtx.DB.Create(&model.PmsProductSpu{
		ID:           901,
		Name:         "旧商品",
		ProductSn:    "SPU-901",
		CategoryID:   11,
		CategoryName: "测试分类",
		BrandID:      21,
		BrandName:    "测试品牌",
		Unit:         "件",
		MainPic:      "old.png",
		PriceRange:   "88.00",
		Stock:        3,
		LowStock:     1,
		CreateBy:     1,
		CreateTime:   now,
	}).Error; err != nil {
		t.Fatalf("seed spu failed: %v", err)
	}
	if err := svcCtx.DB.Exec(`UPDATE pms_product_spu SET platform_id=?, tenant_id=?, merchant_id=? WHERE id=?`, scope.PlatformID, scope.TenantID, scope.MerchantID, 901).Error; err != nil {
		t.Fatalf("seed spu scope failed: %v", err)
	}
	if err := svcCtx.DB.Create(&model.PmsProductSku{
		ID:         902,
		SpuID:      901,
		Name:       "旧SKU",
		SkuCode:    "OLD-SKU",
		Price:      88,
		Stock:      3,
		LowStock:   1,
		SpecData:   `{"颜色":"黑色","尺码":"L"}`,
		CreateBy:   1,
		CreateTime: now,
	}).Error; err != nil {
		t.Fatalf("seed sku failed: %v", err)
	}
	if err := svcCtx.DB.Exec(`UPDATE pms_product_sku SET platform_id=?, tenant_id=?, merchant_id=? WHERE id=?`, scope.PlatformID, scope.TenantID, scope.MerchantID, 902).Error; err != nil {
		t.Fatalf("seed sku scope failed: %v", err)
	}

	logic := NewUpdateProductSpuLogic(context.Background(), svcCtx)
	_, err := logic.UpdateProductSpu(&pmsclient.ProductSpuReq{
		Id:           901,
		Name:         "新商品",
		ProductSn:    "SPU-901",
		CategoryId:   11,
		CategoryName: "测试分类",
		BrandId:      21,
		BrandName:    "测试品牌",
		Unit:         "件",
		MainPic:      "new.png",
		AlbumPics:    "new.png",
		CreateBy:     1002,
		Scope:        &pmsclient.GovernanceScope{ScopeType: scope.ScopeType, PlatformId: scope.PlatformID, TenantId: scope.TenantID, MerchantId: scope.MerchantID},
		MemberPriceList: []*pmsclient.MemberPriceList{{LevelId: 2, Price: 108, LevelName: "白金会员"}},
		ProductLadderList: []*pmsclient.ProductLadderList{{Count: 3, Discount: 85, Price: 119}},
		ProductFullReductionList: []*pmsclient.ProductFullReductionList{{FullPrice: 299, ReducePrice: 30}},
		ProductAttributeValueList: []*pmsclient.ProductAttributeValueList{{ProductAttributeId: 31, AttributeValues: "白色"}},
		SkuStockList: []*pmsclient.SkuStockList{
			{Name: "白色-M", Price: 109, Stock: 4, LowStock: 1, SpecData: `{"颜色":"白色","尺码":"M"}`},
			{Name: "白色-L", Price: 139, Stock: 6, LowStock: 2, SpecData: `{"颜色":"白色","尺码":"L"}`},
		},
	})
	if err != nil {
		t.Fatalf("update product spu failed: %v", err)
	}

	var spu model.PmsProductSpu
	if err := svcCtx.DB.First(&spu, 901).Error; err != nil {
		t.Fatalf("reload spu failed: %v", err)
	}
	if spu.PriceRange != "109.00-139.00" || spu.Stock != 10 || spu.LowStock != 3 {
		t.Fatalf("unexpected updated summary: priceRange=%s stock=%d lowStock=%d", spu.PriceRange, spu.Stock, spu.LowStock)
	}

	var skus []model.PmsProductSku
	if err := svcCtx.DB.Where("spu_id = ?", 901).Order("id asc").Find(&skus).Error; err != nil {
		t.Fatalf("reload skus failed: %v", err)
	}
	if len(skus) != 2 {
		t.Fatalf("expected 2 skus after rebuild, got %d", len(skus))
	}
	for _, sku := range skus {
		if sku.SkuCode == "OLD-SKU" || !strings.HasPrefix(sku.SkuCode, "SPU901-") {
			t.Fatalf("expected deterministic rebuilt sku code, got %s", sku.SkuCode)
		}
	}

	for _, target := range []struct {
		table  string
		column string
		label  string
	}{
		{table: "pms_member_price", column: "product_id", label: "member price"},
		{table: "pms_product_ladder", column: "product_id", label: "product ladder"},
		{table: "pms_product_full_reduction", column: "product_id", label: "full reduction"},
		{table: "pms_product_attribute_value", column: "spu_id", label: "attribute value"},
	} {
		var rows []struct {
			PlatformID int64 `gorm:"column:platform_id"`
			TenantID   int64 `gorm:"column:tenant_id"`
			MerchantID int64 `gorm:"column:merchant_id"`
		}
		if err := svcCtx.DB.Table(target.table).Select("platform_id, tenant_id, merchant_id").Where(target.column+" = ?", 901).Find(&rows).Error; err != nil {
			t.Fatalf("load %s scope failed: %v", target.label, err)
		}
		for _, row := range rows {
			if row.PlatformID != scope.PlatformID || row.TenantID != scope.TenantID || row.MerchantID != scope.MerchantID {
				t.Fatalf("unexpected %s scope: %+v", target.label, row)
			}
		}
	}
}

func TestUpdateVerifyStatusCreatesReviewRecordForDraft(t *testing.T) {
	svcCtx, scope := newProductSpuDraftTestSvc(t)
	verifyModel := &mockProductVertifyRecordModel{}
	svcCtx.ProductVertifyRecordModel = verifyModel

	now := time.Now()
	if err := svcCtx.DB.Create(&model.PmsProductSpu{
		ID:           1201,
		Name:         "待送审商品",
		ProductSn:    "SPU-1201",
		CategoryID:   11,
		CategoryName: "测试分类",
		BrandID:      21,
		BrandName:    "测试品牌",
		Unit:         "件",
		MainPic:      "draft.png",
		PriceRange:   "99.00",
		Stock:        8,
		LowStock:     2,
		VerifyStatus: 0,
		CreateBy:     1001,
		CreateTime:   now,
	}).Error; err != nil {
		t.Fatalf("seed draft spu failed: %v", err)
	}
	if err := svcCtx.DB.Exec(`UPDATE pms_product_spu SET platform_id=?, tenant_id=?, merchant_id=? WHERE id=?`, scope.PlatformID, scope.TenantID, scope.MerchantID, 1201).Error; err != nil {
		t.Fatalf("seed draft spu scope failed: %v", err)
	}

	logic := NewUpdateVerifyStatusLogic(context.Background(), svcCtx)
	_, err := logic.UpdateVerifyStatus(&pmsclient.UpdateProductSpuStatusReq{
		Ids:       []int64{1201},
		Status:    1,
		UpdateBy:  2001,
		ReviewMan: "reviewer",
		Detail:    "草稿满足最小送审条件，进入审核链路",
		Scope: &pmsclient.GovernanceScope{
			ScopeType:  scope.ScopeType,
			PlatformId: scope.PlatformID,
			TenantId:   scope.TenantID,
			MerchantId: scope.MerchantID,
		},
	})
	if err != nil {
		t.Fatalf("UpdateVerifyStatus failed: %v", err)
	}

	var spu model.PmsProductSpu
	if err := svcCtx.DB.First(&spu, 1201).Error; err != nil {
		t.Fatalf("reload spu failed: %v", err)
	}
	if spu.VerifyStatus != 1 {
		t.Fatalf("expected verify status 1, got %d", spu.VerifyStatus)
	}
	if len(verifyModel.inserted) != 1 {
		t.Fatalf("expected 1 verify record, got %d", len(verifyModel.inserted))
	}
	record := verifyModel.inserted[0]
	if record.ProductId != 1201 || record.Status != 1 || record.ReviewMan != "reviewer" {
		t.Fatalf("unexpected verify record: %+v", record)
	}
	if record.Detail != "草稿满足最小送审条件，进入审核链路" {
		t.Fatalf("expected review detail to persist, got %q", record.Detail)
	}
}
