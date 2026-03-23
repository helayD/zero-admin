package common

import (
	"context"
	"fmt"
	"sort"
	"strings"

	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	"github.com/feihua/zero-admin/rpc/pms/gen/model"
	"gorm.io/gorm"
)

type persistedSkuDraftRow struct {
	ID             int64   `gorm:"column:id"`
	SpuID          int64   `gorm:"column:spu_id"`
	Name           string  `gorm:"column:name"`
	SkuCode        string  `gorm:"column:sku_code"`
	Price          float64 `gorm:"column:price"`
	PromotionPrice float64 `gorm:"column:promotion_price"`
	Stock          int32   `gorm:"column:stock"`
	LowStock       int32   `gorm:"column:low_stock"`
	SpecData       string  `gorm:"column:spec_data"`
}

func ComposeSkuValidationSet(ctx context.Context, db *gorm.DB, spuID int64, incoming []ProductDraftSku, excludeIDs []int64) ([]ProductDraftSku, error) {
	if spuID <= 0 {
		return nil, fmt.Errorf("缺少有效商品SPU")
	}
	excluded := make(map[int64]struct{}, len(excludeIDs))
	for _, id := range excludeIDs {
		if id > 0 {
			excluded[id] = struct{}{}
		}
	}

	var rows []persistedSkuDraftRow
	if err := db.WithContext(ctx).
		Table("pms_product_sku").
		Select("id, spu_id, name, sku_code, price, promotion_price, stock, low_stock, spec_data").
		Where("spu_id = ? AND is_deleted = 0", spuID).
		Find(&rows).Error; err != nil {
		return nil, err
	}

	combined := make([]ProductDraftSku, 0, len(rows)+len(incoming))
	for _, row := range rows {
		if _, ok := excluded[row.ID]; ok {
			continue
		}
		combined = append(combined, ProductDraftSku{
			ID:             row.ID,
			SpuID:          row.SpuID,
			Name:           row.Name,
			SkuCode:        row.SkuCode,
			Price:          float32(row.Price),
			PromotionPrice: float32(row.PromotionPrice),
			Stock:          row.Stock,
			LowStock:       row.LowStock,
			SpecData:       row.SpecData,
		})
	}
	combined = append(combined, incoming...)
	return combined, nil
}

func RefreshSpuDraftSummary(ctx context.Context, db *gorm.DB, current pkgscope.GovernanceScope, spuID int64) error {
	var rows []persistedSkuDraftRow
	if err := db.WithContext(ctx).
		Table("pms_product_sku").
		Select("id, spu_id, name, sku_code, price, promotion_price, stock, low_stock, spec_data").
		Where("spu_id = ? AND is_deleted = 0", spuID).
		Find(&rows).Error; err != nil {
		return err
	}

	summary := &DraftSummary{PriceRange: "", TotalStock: 0, LowStock: 0}
	if len(rows) > 0 {
		prices := make([]float64, 0, len(rows))
		for _, row := range rows {
			prices = append(prices, row.Price)
			summary.TotalStock += row.Stock
			summary.LowStock += row.LowStock
		}
		sort.Float64s(prices)
		summary.PriceRange = fmt.Sprintf("%.2f", prices[0])
		if prices[0] != prices[len(prices)-1] {
			summary.PriceRange = fmt.Sprintf("%.2f-%.2f", prices[0], prices[len(prices)-1])
		}
	}

	if err := db.WithContext(ctx).Table("pms_product_spu").Where("id = ?", spuID).Updates(map[string]interface{}{
		"price_range": summary.PriceRange,
		"stock":       summary.TotalStock,
		"low_stock":   summary.LowStock,
	}).Error; err != nil {
		return err
	}
	return ApplyProductScope(ctx, db, spuID, current)
}

func EnsureSkuCode(ctx context.Context, db *gorm.DB, current pkgscope.GovernanceScope, spuID int64, rawCode, specData, name string, ignoreIDs map[int64]struct{}) (string, error) {
	skuCode := strings.TrimSpace(rawCode)
	if skuCode == "" {
		generated, err := GenerateSkuCode(spuID, specData, name)
		if err != nil {
			return "", err
		}
		skuCode = generated
	}
	if err := ensureSkuCodeAvailable(ctx, db, current, skuCode, ignoreIDs, 0); err != nil {
		return "", err
	}
	return skuCode, nil
}

func GenerateSkuCode(spuID int64, specData, name string) (string, error) {
	if spuID <= 0 {
		return "", fmt.Errorf("缺少有效商品SPU")
	}
	canonicalSpec, err := normalizeSpecData(specData)
	if err != nil {
		return "", err
	}
	base := canonicalSpec
	if strings.TrimSpace(base) == "" {
		base = strings.TrimSpace(name)
	}
	replacer := strings.NewReplacer("|", "-", "=", "-", " ", "", "/", "-", "_", "-", ".", "-", ",", "-")
	normalized := strings.ToUpper(replacer.Replace(base))
	parts := strings.FieldsFunc(normalized, func(r rune) bool {
		return !(r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '-')
	})
	normalized = strings.Join(parts, "-")
	normalized = strings.Trim(normalized, "-")
	if normalized == "" {
		normalized = fmt.Sprintf("SKU-%d", spuID)
	}
	code := fmt.Sprintf("SPU%d-%s", spuID, normalized)
	if len(code) > 50 {
		code = code[:50]
		code = strings.TrimRight(code, "-")
	}
	return code, nil
}

func BuildSkuModelFromDraft(spuID int64, operatorID int64, draft ProductDraftSku, base *model.PmsProductSku) *model.PmsProductSku {
	item := &model.PmsProductSku{
		ID:             draft.ID,
		SpuID:          spuID,
		Name:           draft.Name,
		SkuCode:        draft.SkuCode,
		Price:          float64(draft.Price),
		PromotionPrice: float64(draft.PromotionPrice),
		Stock:          draft.Stock,
		LowStock:       draft.LowStock,
		SpecData:       draft.SpecData,
	}
	if base != nil {
		item.MainPic = base.MainPic
		item.AlbumPics = base.AlbumPics
		item.Weight = base.Weight
		item.PublishStatus = base.PublishStatus
		item.VerifyStatus = base.VerifyStatus
		item.Sort = base.Sort
		item.Sales = base.Sales
		item.CreateBy = base.CreateBy
		item.CreateTime = base.CreateTime
		item.IsDeleted = base.IsDeleted
		item.PromotionStartTime = base.PromotionStartTime
		item.PromotionEndTime = base.PromotionEndTime
		item.UpdateBy = base.UpdateBy
		item.UpdateTime = base.UpdateTime
	}
	if item.ID == 0 {
		item.CreateBy = operatorID
	}
	if operatorID > 0 {
		item.UpdateBy = &operatorID
	}
	return item
}
