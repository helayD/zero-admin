package digitalcardmint

import (
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

type EnsurePaidOrderPurchaseAssetsInput struct {
	OrderID      int64
	MemberID     int64
	PlatformID   int64
	TenantID     int64
	MerchantID   int64
	EventID      string
	TraceID      string
	OperatorType string
}

type EnsurePaidOrderPurchaseAssetsResult struct {
	ProcessedCount   int
	AssetInstanceIDs []int64
}

type paidOrderItemRow struct {
	ID      int64  `gorm:"column:id"`
	OrderID int64  `gorm:"column:order_id"`
	SkuID   int64  `gorm:"column:sku_id"`
	SkuName string `gorm:"column:sku_name"`
}

type fulfillmentProductRow struct {
	SkuID             int64  `gorm:"column:sku_id"`
	ProductID         int64  `gorm:"column:product_id"`
	FulfillmentMode   string `gorm:"column:fulfillment_mode"`
	FulfillmentRuleID int64  `gorm:"column:fulfillment_rule_id"`
	PlatformID        int64  `gorm:"column:platform_id"`
	TenantID          int64  `gorm:"column:tenant_id"`
	MerchantID        int64  `gorm:"column:merchant_id"`
}

func (s *Service) EnsurePaidOrderPurchaseAssets(ctx context.Context, input EnsurePaidOrderPurchaseAssetsInput) (*EnsurePaidOrderPurchaseAssetsResult, error) {
	result := &EnsurePaidOrderPurchaseAssetsResult{}
	if s == nil || s.DB == nil || input.OrderID <= 0 {
		return result, nil
	}

	items, err := loadPaidOrderItems(ctx, s.DB, input.OrderID)
	if err != nil {
		return nil, err
	}
	if len(items) == 0 {
		return result, nil
	}

	skuIDs := uniquePositiveSkuIDs(items)
	if len(skuIDs) == 0 {
		return result, nil
	}

	productBySkuID, err := loadFulfillmentProductsBySkuID(ctx, s.DB, skuIDs)
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		product := productBySkuID[item.SkuID]
		if product == nil || strings.TrimSpace(product.FulfillmentMode) != "digital_asset" {
			continue
		}
		asset, ensureErr := s.EnsureOrderPurchaseAsset(ctx, EnsureOrderPurchaseAssetInput{
			OrderID:           input.OrderID,
			OrderItemID:       item.ID,
			ProductID:         product.ProductID,
			SkuID:             item.SkuID,
			MemberID:          input.MemberID,
			FulfillmentRuleID: product.FulfillmentRuleID,
			PlatformID:        firstPositive(input.PlatformID, product.PlatformID),
			TenantID:          firstPositive(input.TenantID, product.TenantID),
			MerchantID:        firstPositive(input.MerchantID, product.MerchantID),
			RequestID:         input.EventID,
			TraceID:           input.TraceID,
			OperatorType:      input.OperatorType,
		})
		if ensureErr != nil {
			return nil, fmt.Errorf("\u8ba2\u5355\u660e\u7ec6[%d](skuId=%d, productId=%d, ruleId=%d, scope=p%d/t%d/m%d)\u63d0\u8d27\u5361\u5efa\u8d26\u5931\u8d25: %w", item.ID, item.SkuID, product.ProductID, product.FulfillmentRuleID, firstPositive(input.PlatformID, product.PlatformID), firstPositive(input.TenantID, product.TenantID), firstPositive(input.MerchantID, product.MerchantID), ensureErr)
		}
		result.ProcessedCount++
		if asset != nil && asset.AssetInstanceID > 0 {
			result.AssetInstanceIDs = append(result.AssetInstanceIDs, asset.AssetInstanceID)
		}
	}
	return result, nil
}

func loadPaidOrderItems(ctx context.Context, db *gorm.DB, orderID int64) ([]paidOrderItemRow, error) {
	var items []paidOrderItemRow
	err := db.WithContext(ctx).
		Table("oms_order_item").
		Select("id, order_id, sku_id, sku_name").
		Where("order_id = ? AND is_deleted = 0", orderID).
		Find(&items).Error
	return items, err
}

func uniquePositiveSkuIDs(items []paidOrderItemRow) []int64 {
	skuIDs := make([]int64, 0, len(items))
	seen := make(map[int64]struct{}, len(items))
	for _, item := range items {
		if item.SkuID <= 0 {
			continue
		}
		if _, ok := seen[item.SkuID]; ok {
			continue
		}
		seen[item.SkuID] = struct{}{}
		skuIDs = append(skuIDs, item.SkuID)
	}
	return skuIDs
}

func loadFulfillmentProductsBySkuID(ctx context.Context, db *gorm.DB, skuIDs []int64) (map[int64]*fulfillmentProductRow, error) {
	var rows []fulfillmentProductRow
	err := db.WithContext(ctx).
		Table("pms_product_sku AS sku").
		Joins("JOIN pms_product_spu AS spu ON spu.id = sku.spu_id AND spu.is_deleted = 0").
		Select("sku.id AS sku_id, spu.id AS product_id, COALESCE(NULLIF(sku.fulfillment_mode, ''), spu.fulfillment_mode) AS fulfillment_mode, COALESCE(NULLIF(sku.fulfillment_rule_id, 0), spu.fulfillment_rule_id) AS fulfillment_rule_id, spu.platform_id, spu.tenant_id, spu.merchant_id").
		Where("sku.id IN ? AND sku.is_deleted = 0", skuIDs).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	result := make(map[int64]*fulfillmentProductRow, len(rows))
	for index := range rows {
		row := rows[index]
		result[row.SkuID] = &row
	}
	return result, nil
}

func firstPositive(values ...int64) int64 {
	for _, value := range values {
		if value > 0 {
			return value
		}
	}
	return 0
}
