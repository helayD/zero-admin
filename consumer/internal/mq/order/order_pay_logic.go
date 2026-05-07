package order

import (
	"context"
	"fmt"
	"strings"

	"github.com/bytedance/sonic"
	"github.com/feihua/zero-admin/consumer/internal/mq/member"
	"github.com/feihua/zero-admin/pkg/digitalcardmint"
	"github.com/feihua/zero-admin/rpc/ums/client/membermessageservice"
	"github.com/zeromicro/go-zero/core/logc"
	"gorm.io/gorm"
)

// OrderPay 订单支付成功消息处理
// Story 8-1 Fix #1: 新增支付成功消息消费者
// Story 10.6: 增加履约分叉逻辑，数字资产商品触发 SMS 发卡
func OrderPay(ctx context.Context, body []byte, memberMsgService membermessageservice.MemberMessageService, cardMintService *digitalcardmint.Service, db *gorm.DB) error {
	var payload EventPayload
	if err := sonic.Unmarshal(body, &payload); err != nil {
		logc.Errorf(ctx, "OrderPay 反序列化失败: %v", err)
		return err
	}
	payload.NormalizeLegacyPaymentPayload()

	ctx = payload.ToContext(ctx)
	LogWithEventPayload(ctx, "OrderPay 收到订单支付成功事件, entityId=%d, action=%s", payload.EntityID, payload.Action)

	orderNo := ""
	if v, ok := payload.Data["orderNo"]; ok {
		orderNo, _ = v.(string)
	}

	// 1. 发送支付成功消息通知
	msgEvent := member.MemberMessageEvent{
		MemberId:    payload.ActorID,
		MessageType: member.MessageTypePayment,
		Title:       "支付成功",
		Content:     fmt.Sprintf("您的订单（%s）已支付成功，感谢您的购买！", orderNo),
		LinkType:    "order",
		LinkId:      payload.EntityIDToString(),
		PlatformId:  payload.PlatformID,
		TenantId:    payload.TenantID,
		MerchantId:  payload.MerchantID,
	}

	body2, err := sonic.Marshal(msgEvent)
	if err != nil {
		LogWithEventPayload(ctx, "OrderPay 序列化消息事件失败: %v", err)
		return nil
	}

	if err := member.CreateMemberMessage(ctx, body2, memberMsgService); err != nil {
		LogWithEventPayload(ctx, "OrderPay 发送支付成功消息失败: %v", err)
		// 不返回错误，继续处理履约分叉
	}

	// 2. 履约分叉逻辑（Story 10.6）
	if err := processPaidOrderDigitalAssets(ctx, &payload, cardMintService, db); err != nil {
		LogWithEventPayload(ctx, "OrderPay 数字资产履约分叉失败, orderId=%d, err=%v", payload.EntityID, err)
		return err
	}
	LogWithEventPayload(ctx, "OrderPay 履约分叉处理完成, orderId=%d", payload.EntityID)

	return nil
}

type paidOrderItemRow struct {
	ID      int64  `gorm:"column:id"`
	OrderID int64  `gorm:"column:order_id"`
	SkuID   int64  `gorm:"column:sku_id"`
	SkuName string `gorm:"column:sku_name"`
}

func processPaidOrderDigitalAssets(ctx context.Context, payload *EventPayload, cardMintService *digitalcardmint.Service, db *gorm.DB) error {
	if payload == nil || payload.EntityID <= 0 || db == nil || cardMintService == nil {
		return nil
	}

	var items []paidOrderItemRow
	if err := db.WithContext(ctx).
		Table("oms_order_item").
		Select("id, order_id, sku_id, sku_name").
		Where("order_id = ? AND is_deleted = 0", payload.EntityID).
		Find(&items).Error; err != nil {
		return err
	}
	if len(items) == 0 {
		return nil
	}

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
	if len(skuIDs) == 0 {
		return nil
	}

	productBySkuID, err := loadFulfillmentProductsBySkuID(ctx, db, skuIDs)
	if err != nil {
		return err
	}

	for _, item := range items {
		product := productBySkuID[item.SkuID]
		if product == nil || strings.TrimSpace(product.FulfillmentMode) != "digital_asset" {
			continue
		}
		_, err = cardMintService.EnsureOrderPurchaseAsset(ctx, digitalcardmint.EnsureOrderPurchaseAssetInput{
			OrderID:           payload.EntityID,
			OrderItemID:       item.ID,
			ProductID:         product.ProductID,
			SkuID:             item.SkuID,
			MemberID:          payload.ActorID,
			FulfillmentRuleID: product.FulfillmentRuleID,
			PlatformID:        firstPositive(payload.PlatformID, product.PlatformID),
			TenantID:          firstPositive(payload.TenantID, product.TenantID),
			MerchantID:        firstPositive(payload.MerchantID, product.MerchantID),
			RequestID:         payload.EventID,
			TraceID:           payload.TraceID,
			OperatorType:      "system",
		})
		if err != nil {
			return fmt.Errorf("订单明细[%d]数字资产建账失败: %w", item.ID, err)
		}
	}
	return nil
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

func loadFulfillmentProductsBySkuID(ctx context.Context, db *gorm.DB, skuIDs []int64) (map[int64]*fulfillmentProductRow, error) {
	var rows []fulfillmentProductRow
	err := db.WithContext(ctx).
		Table("pms_product_sku AS sku").
		Joins("JOIN pms_product_spu AS spu ON spu.id = sku.spu_id AND spu.is_deleted = 0").
		Select("sku.id AS sku_id, spu.id AS product_id, COALESCE(NULLIF(sku.fulfillment_mode, ''), spu.fulfillment_mode) AS fulfillment_mode, COALESCE(sku.fulfillment_rule_id, spu.fulfillment_rule_id) AS fulfillment_rule_id, spu.platform_id, spu.tenant_id, spu.merchant_id").
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
