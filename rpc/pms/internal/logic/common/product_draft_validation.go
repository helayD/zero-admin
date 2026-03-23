package common

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	"github.com/feihua/zero-admin/rpc/pms/pmsclient"
	"gorm.io/gorm"
)

type scopedStatusRow struct {
	ID         int64 `gorm:"column:id"`
	Status     int32 `gorm:"column:status"`
	IsEnabled  int32 `gorm:"column:is_enabled"`
	PlatformID int64 `gorm:"column:platform_id"`
	TenantID   int64 `gorm:"column:tenant_id"`
	MerchantID int64 `gorm:"column:merchant_id"`
}

type DraftSummary struct {
	PriceRange string
	TotalStock int32
	LowStock   int32
}

type ProductDraftSku struct {
	ID             int64
	SpuID          int64
	Name           string
	SkuCode        string
	Price          float32
	PromotionPrice float32
	Stock          int32
	LowStock       int32
	SpecData       string
}

func EnsureScopedBrandExists(ctx context.Context, db *gorm.DB, current pkgscope.GovernanceScope, brandID int64, message string) error {
	if brandID <= 0 {
		return errors.New("缺少有效商品品牌ID")
	}
	var row scopedStatusRow
	if err := db.WithContext(ctx).Table("pms_product_brand").Select("id, is_enabled, platform_id, tenant_id, merchant_id").Where("id = ? AND is_deleted = 0", brandID).Take(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("商品品牌不存在")
		}
		return err
	}
	if !current.SameScope(pkgscope.DefaultScope(row.PlatformID, row.TenantID, row.MerchantID)) {
		return errors.New(message)
	}
	if row.IsEnabled != 1 {
		return errors.New("商品品牌已失效，无法继续建档")
	}
	return nil
}

func EnsureScopedAttributeExists(ctx context.Context, db *gorm.DB, current pkgscope.GovernanceScope, attributeID int64, message string) error {
	if attributeID <= 0 {
		return errors.New("缺少有效商品属性ID")
	}
	var row scopedStatusRow
	if err := db.WithContext(ctx).Table("pms_product_attribute").Select("id, status, platform_id, tenant_id, merchant_id").Where("id = ? AND is_deleted = 0", attributeID).Take(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("商品属性不存在")
		}
		return err
	}
	if !current.SameScope(pkgscope.DefaultScope(row.PlatformID, row.TenantID, row.MerchantID)) {
		return errors.New(message)
	}
	if row.Status != 1 {
		return errors.New("商品属性已失效，无法继续建档")
	}
	return nil
}

func ValidateProductDraft(ctx context.Context, db *gorm.DB, current pkgscope.GovernanceScope, in *pmsclient.ProductSpuReq) (*DraftSummary, error) {
	if in == nil {
		return nil, errors.New("商品建档参数不能为空")
	}
	if strings.TrimSpace(in.Name) == "" {
		return nil, errors.New("商品名称不能为空")
	}
	if strings.TrimSpace(in.ProductSn) == "" {
		return nil, errors.New("商品货号不能为空")
	}
	if strings.TrimSpace(in.MainPic) == "" {
		return nil, errors.New("商品主图不能为空")
	}
	if err := EnsureScopedCategoryExists(ctx, db, current, in.CategoryId, "当前主体无权将商品归属到该商品分类"); err != nil {
		return nil, err
	}
	if err := EnsureScopedBrandExists(ctx, db, current, in.BrandId, "当前主体无权将商品归属到该商品品牌"); err != nil {
		return nil, err
	}

	seenAttributes := make(map[int64]struct{})
	for _, attr := range in.ProductAttributeValueList {
		if err := EnsureScopedAttributeExists(ctx, db, current, attr.ProductAttributeId, "当前主体无权绑定该商品属性"); err != nil {
			return nil, err
		}
		if _, ok := seenAttributes[attr.ProductAttributeId]; ok {
			return nil, fmt.Errorf("商品属性重复填写: %d", attr.ProductAttributeId)
		}
		seenAttributes[attr.ProductAttributeId] = struct{}{}
		if strings.TrimSpace(attr.AttributeValues) == "" {
			return nil, fmt.Errorf("商品属性值不能为空: %d", attr.ProductAttributeId)
		}
	}

	skuList := make([]ProductDraftSku, 0, len(in.SkuStockList))
	for _, sku := range in.SkuStockList {
		skuList = append(skuList, ProductDraftSku{
			SpuID:          in.Id,
			Name:           sku.Name,
			SkuCode:        sku.SkuCode,
			Price:          sku.Price,
			PromotionPrice: sku.PromotionPrice,
			Stock:          sku.Stock,
			LowStock:       sku.LowStock,
			SpecData:       sku.SpecData,
		})
	}
	return ValidateSkuDrafts(ctx, db, current, skuList, in.Id)
}

func ValidateSkuDrafts(ctx context.Context, db *gorm.DB, current pkgscope.GovernanceScope, skuList []ProductDraftSku, ignoreSpuID int64) (*DraftSummary, error) {
	if len(skuList) == 0 {
		return nil, errors.New("至少需要一个有效SKU")
	}
	seenSpec := make(map[string]struct{}, len(skuList))
	seenCode := make(map[string]struct{}, len(skuList))
	ignoreIDs := make(map[int64]struct{}, len(skuList))
	var prices []float32
	var totalStock int32
	var totalLowStock int32
	for idx, sku := range skuList {
		line := idx + 1
		if sku.ID > 0 {
			ignoreIDs[sku.ID] = struct{}{}
		}
		if strings.TrimSpace(sku.Name) == "" {
			return nil, fmt.Errorf("第%d行SKU名称不能为空", line)
		}
		if sku.Price <= 0 {
			return nil, fmt.Errorf("第%d行SKU价格必须大于0", line)
		}
		if sku.PromotionPrice < 0 {
			return nil, fmt.Errorf("第%d行SKU促销价不能小于0", line)
		}
		if sku.PromotionPrice > 0 && sku.PromotionPrice > sku.Price {
			return nil, fmt.Errorf("第%d行SKU促销价不能高于销售价", line)
		}
		if sku.Stock < 0 {
			return nil, fmt.Errorf("第%d行SKU库存不能小于0", line)
		}
		if sku.LowStock < 0 {
			return nil, fmt.Errorf("第%d行SKU预警库存不能小于0", line)
		}
		if sku.LowStock > sku.Stock {
			return nil, fmt.Errorf("第%d行SKU预警库存不能大于可用库存", line)
		}
		canonicalSpec, err := normalizeSpecData(sku.SpecData)
		if err != nil {
			return nil, fmt.Errorf("第%d行SKU规格组合非法: %w", line, err)
		}
		if _, ok := seenSpec[canonicalSpec]; ok {
			return nil, fmt.Errorf("第%d行SKU规格组合重复", line)
		}
		seenSpec[canonicalSpec] = struct{}{}

		skuCode := strings.TrimSpace(sku.SkuCode)
		if skuCode != "" {
			if _, ok := seenCode[skuCode]; ok {
				return nil, fmt.Errorf("第%d行SKU编码重复", line)
			}
			seenCode[skuCode] = struct{}{}
			if err := ensureSkuCodeAvailable(ctx, db, current, skuCode, ignoreIDs, ignoreSpuID); err != nil {
				return nil, fmt.Errorf("第%d行SKU编码冲突: %w", line, err)
			}
		}

		prices = append(prices, sku.Price)
		totalStock += sku.Stock
		totalLowStock += sku.LowStock
	}
	if totalStock <= 0 {
		return nil, errors.New("商品草稿不可保存：至少一个SKU需要具备有效库存")
	}
	sort.Slice(prices, func(i, j int) bool { return prices[i] < prices[j] })
	priceRange := fmt.Sprintf("%.2f", prices[0])
	if len(prices) > 1 && prices[len(prices)-1] != prices[0] {
		priceRange = fmt.Sprintf("%.2f-%.2f", prices[0], prices[len(prices)-1])
	}
	return &DraftSummary{PriceRange: priceRange, TotalStock: totalStock, LowStock: totalLowStock}, nil
}

func normalizeSpecData(raw string) (string, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return "", errors.New("缺少规格主键")
	}
	var spec map[string]interface{}
	if err := json.Unmarshal([]byte(trimmed), &spec); err != nil {
		return "", errors.New("规格数据不是合法JSON")
	}
	if len(spec) == 0 {
		return "", errors.New("规格数据不能为空对象")
	}
	keys := make([]string, 0, len(spec))
	for k, v := range spec {
		key := strings.TrimSpace(k)
		value := strings.TrimSpace(fmt.Sprint(v))
		if key == "" || value == "" || value == "<nil>" {
			return "", errors.New("规格键和值不能为空")
		}
		keys = append(keys, key+"="+value)
	}
	sort.Strings(keys)
	return strings.Join(keys, "|"), nil
}

func ensureSkuCodeAvailable(ctx context.Context, db *gorm.DB, current pkgscope.GovernanceScope, skuCode string, ignoreIDs map[int64]struct{}, ignoreSpuID int64) error {
	var rows []struct {
		ID         int64 `gorm:"column:id"`
		SpuID      int64 `gorm:"column:spu_id"`
		PlatformID int64 `gorm:"column:platform_id"`
		TenantID   int64 `gorm:"column:tenant_id"`
		MerchantID int64 `gorm:"column:merchant_id"`
	}
	if err := db.WithContext(ctx).Table("pms_product_sku").Select("id, spu_id, platform_id, tenant_id, merchant_id").Where("sku_code = ? AND is_deleted = 0", skuCode).Find(&rows).Error; err != nil {
		return err
	}
	for _, row := range rows {
		if _, ok := ignoreIDs[row.ID]; ok {
			continue
		}
		if ignoreSpuID > 0 && row.SpuID == ignoreSpuID {
			continue
		}
		if current.SameScope(pkgscope.DefaultScope(row.PlatformID, row.TenantID, row.MerchantID)) {
			return errors.New("当前主体下已存在相同SKU编码")
		}
	}
	return nil
}
