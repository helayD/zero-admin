package jobs

import (
	"context"
	"fmt"
	"time"

	"github.com/feihua/zero-admin/pkg/digitalcardmint"
	"github.com/zeromicro/go-zero/core/logc"
	"gorm.io/gorm"
)

// ReconcileOrderCardFulfillment 对账扫描任务
// 扫描已支付但 digital_asset 模式订单明细在超时后仍未生成对应卡片实例
// 扫描已生成卡片实例但订单状态与资产状态不一致（如订单已退款但卡片仍然可提货）
func ReconcileOrderCardFulfillment(ctx context.Context, db *gorm.DB, cardMintService *digitalcardmint.Service) {
	if db == nil || cardMintService == nil {
		logc.Errorf(ctx, "ReconcileOrderCardFulfillment: db or cardMintService is nil")
		return
	}

	logc.Infof(ctx, "ReconcileOrderCardFulfillment: 开始对账扫描")

	// 1. 扫描应发卡未发卡的订单明细
	missingCount, err := reconcileMissingCardAssets(ctx, db, cardMintService)
	if err != nil {
		logc.Errorf(ctx, "ReconcileOrderCardFulfillment: 扫描应发卡未发卡失败: %v", err)
	}

	// 2. 扫描订单已退款但卡片仍然可提货的情况
	refundMismatchCount, err := reconcileRefundedOrderCards(ctx, db, cardMintService)
	if err != nil {
		logc.Errorf(ctx, "ReconcileOrderCardFulfillment: 扫描退款订单卡片状态失败: %v", err)
	}

	logc.Infof(ctx, "ReconcileOrderCardFulfillment: 对账扫描完成, 应发卡未发卡: %d, 退款订单卡片状态不一致: %d", missingCount, refundMismatchCount)
}

// reconcileMissingCardAssets 扫描应发卡未发卡的订单明细
// 条件：订单已支付 + 订单明细商品为 digital_asset 模式 + 超过 5 分钟仍未生成卡片实例
//
// 注意：oms_order_main 上没有 pay_status / member_id 列，order_status=2 已经表示
// "已支付"（参考 oms_order_main.gen.go 的列注释）。会员 ID 在主表上叫 user_id。
func reconcileMissingCardAssets(ctx context.Context, db *gorm.DB, cardMintService *digitalcardmint.Service) (int64, error) {
	type missingAssetRow struct {
		OrderID           int64     `gorm:"column:order_id"`
		OrderItemID       int64     `gorm:"column:order_item_id"`
		SkuID             int64     `gorm:"column:sku_id"`
		SkuName           string    `gorm:"column:sku_name"`
		FulfillmentMode   string    `gorm:"column:fulfillment_mode"`
		FulfillmentRuleID int64     `gorm:"column:fulfillment_rule_id"`
		PlatformID        int64     `gorm:"column:platform_id"`
		TenantID          int64     `gorm:"column:tenant_id"`
		MerchantID        int64     `gorm:"column:merchant_id"`
		MemberID          int64     `gorm:"column:member_id"`
		PayTime           time.Time `gorm:"column:pay_time"`
	}

	// 查询已支付超过 5 分钟、digital_asset 模式、但没有对应卡片实例的订单明细
	var rows []missingAssetRow
	err := db.WithContext(ctx).
		Table("oms_order_item AS item").
		Joins("JOIN oms_order_main AS main ON main.id = item.order_id AND main.is_deleted = 0").
		Joins("JOIN pms_product_sku AS sku ON sku.id = item.sku_id AND sku.is_deleted = 0").
		Joins("JOIN pms_product_spu AS spu ON spu.id = sku.spu_id AND spu.is_deleted = 0").
		Select(`
			item.order_id,
			item.id AS order_item_id,
			item.sku_id,
			item.sku_name,
			COALESCE(NULLIF(sku.fulfillment_mode, ''), spu.fulfillment_mode) AS fulfillment_mode,
			COALESCE(NULLIF(sku.fulfillment_rule_id, 0), spu.fulfillment_rule_id) AS fulfillment_rule_id,
			spu.platform_id,
			spu.tenant_id,
			spu.merchant_id,
			main.user_id AS member_id,
			main.pay_time
		`).
		Where(`main.order_status = 2 AND main.is_deleted = 0`).
		Where(`COALESCE(NULLIF(sku.fulfillment_mode, ''), spu.fulfillment_mode) = 'digital_asset'`).
		Where(`main.pay_time IS NOT NULL AND main.pay_time < ?`, time.Now().Add(-5*time.Minute)).
		Where(`NOT EXISTS (
			SELECT 1 FROM sms_card_instance ci
			WHERE ci.source_type = 'purchase' AND ci.source_id = item.id AND ci.is_deleted = 0
		)`).
		Find(&rows).Error
	if err != nil {
		return 0, fmt.Errorf("查询应发卡未发卡订单明细失败: %w", err)
	}

	if len(rows) == 0 {
		logc.Infof(ctx, "reconcileMissingCardAssets: 未发现应发卡未发卡的订单明细")
		return 0, nil
	}

	logc.Infof(ctx, "reconcileMissingCardAssets: 发现 %d 条应发卡未发卡的订单明细", len(rows))

	// 尝试补发卡
	var fixedCount int64
	for _, row := range rows {
		logc.Infof(ctx, "reconcileMissingCardAssets: 尝试补发卡, orderId=%d, orderItemId=%d, skuId=%d", row.OrderID, row.OrderItemID, row.SkuID)

		result, err := cardMintService.EnsureOrderPurchaseAsset(ctx, digitalcardmint.EnsureOrderPurchaseAssetInput{
			OrderID:           row.OrderID,
			OrderItemID:       row.OrderItemID,
			ProductID:         0, // 无法直接获取，但不影响发卡
			SkuID:             row.SkuID,
			MemberID:          row.MemberID,
			FulfillmentRuleID: row.FulfillmentRuleID,
			PlatformID:        row.PlatformID,
			TenantID:          row.TenantID,
			MerchantID:        row.MerchantID,
			RequestID:         fmt.Sprintf("reconcile-%d-%d", row.OrderID, row.OrderItemID),
			TraceID:           fmt.Sprintf("reconcile-%d-%d", row.OrderID, row.OrderItemID),
			OperatorType:      "reconcile",
		})
		if err != nil {
			logc.Errorf(ctx, "reconcileMissingCardAssets: 补发卡失败, orderId=%d, orderItemId=%d, err=%v", row.OrderID, row.OrderItemID, err)
			continue
		}

		logc.Infof(ctx, "reconcileMissingCardAssets: 补发卡成功, orderId=%d, orderItemId=%d, assetInstanceId=%d, assetNo=%s", row.OrderID, row.OrderItemID, result.AssetInstanceID, result.AssetNo)
		fixedCount++
	}

	return fixedCount, nil
}

// reconcileRefundedOrderCards 扫描订单已退款但卡片仍然可提货的情况，并按照
// 资产上的 refund_policy 调用 HandleRefundCardDispose 真正执行冻结/回收/人工复核。
//
// 注意：oms_order_main 上没有 pay_status 列；订单状态 6 才表示"已退款"
// （参考 oms_order_main.gen.go 的列注释 1-待支付,2-已支付,3-已发货,4-已完成,5-已取消,6-已退款,7-售后中）。
func reconcileRefundedOrderCards(ctx context.Context, db *gorm.DB, cardMintService *digitalcardmint.Service) (int64, error) {
	type refundMismatchRow struct {
		OrderID          int64  `gorm:"column:order_id"`
		OrderItemID      int64  `gorm:"column:order_item_id"`
		AssetInstanceID  int64  `gorm:"column:asset_instance_id"`
		AssetNo          string `gorm:"column:asset_no"`
		MintStatus       string `gorm:"column:mint_status"`
		ComplianceStatus string `gorm:"column:compliance_status"`
		RefundPolicy     string `gorm:"column:refund_policy"`
	}

	// 查询订单已退款、但卡片仍然处于可提货状态的情况
	var rows []refundMismatchRow
	err := db.WithContext(ctx).
		Table("oms_order_item AS item").
		Joins("JOIN oms_order_main AS main ON main.id = item.order_id AND main.is_deleted = 0").
		Joins("JOIN sms_card_instance AS ci ON ci.source_type = 'purchase' AND ci.source_id = item.id AND ci.is_deleted = 0").
		Select(`
			item.order_id,
			item.id AS order_item_id,
			ci.id AS asset_instance_id,
			ci.asset_no,
			ci.mint_status,
			ci.compliance_status,
			ci.refund_policy
		`).
		Where(`main.order_status = 6 AND main.is_deleted = 0`).
		Where(`ci.compliance_status NOT IN (?, ?, ?)`,
			digitalcardmint.ComplianceStatusFrozen,
			digitalcardmint.ComplianceStatusRecycled,
			digitalcardmint.ComplianceStatusManualReview).
		Find(&rows).Error
	if err != nil {
		return 0, fmt.Errorf("查询退款订单卡片状态不一致失败: %w", err)
	}

	if len(rows) == 0 {
		logc.Infof(ctx, "reconcileRefundedOrderCards: 未发现退款订单卡片状态不一致的情况")
		return 0, nil
	}

	logc.Infof(ctx, "reconcileRefundedOrderCards: 发现 %d 条退款订单卡片状态不一致", len(rows))

	if cardMintService == nil {
		// 没有处置服务时退化为只记录差异，避免阻塞调度
		for _, row := range rows {
			logc.Infof(ctx, "reconcileRefundedOrderCards(missing service): orderId=%d, orderItemId=%d, assetInstanceId=%d, assetNo=%s, refundPolicy=%s",
				row.OrderID, row.OrderItemID, row.AssetInstanceID, row.AssetNo, row.RefundPolicy)
		}
		return int64(len(rows)), nil
	}

	// 真正按 refund_policy 触发卡片处置（幂等，已 frozen/recycled 不重复处置）
	var disposedCount int64
	for _, row := range rows {
		policy := row.RefundPolicy
		if policy == "" {
			// 老数据 refund_policy 缺失时，按"冻结"做最保守兜底：避免可提货卡片留在已退款订单上
			policy = "freeze_card"
		}

		result, disposeErr := cardMintService.HandleRefundCardDispose(ctx, digitalcardmint.RefundCardDisposeInput{
			OrderID:      row.OrderID,
			OrderItemID:  row.OrderItemID,
			RefundPolicy: policy,
			Reason:       "对账兜底：订单已退款但卡片仍可提货",
			TraceID:      fmt.Sprintf("reconcile-refund-%d-%d", row.OrderID, row.OrderItemID),
		})
		if disposeErr != nil {
			logc.Errorf(ctx, "reconcileRefundedOrderCards: 处置失败, orderId=%d, orderItemId=%d, policy=%s, err=%v",
				row.OrderID, row.OrderItemID, policy, disposeErr)
			continue
		}
		if result != nil {
			logc.Infof(ctx, "reconcileRefundedOrderCards: 处置成功, orderId=%d, orderItemId=%d, assetInstanceId=%d, action=%s, complianceStatus=%s",
				row.OrderID, row.OrderItemID, result.AssetInstanceID, result.RefundAction, result.ComplianceStatus)
		}
		disposedCount++
	}

	return disposedCount, nil
}
