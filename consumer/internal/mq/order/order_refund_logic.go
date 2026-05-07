package order

import (
	"context"
	"fmt"

	"github.com/bytedance/sonic"
	"github.com/feihua/zero-admin/pkg/digitalcardmint"
	"github.com/zeromicro/go-zero/core/logc"
	"gorm.io/gorm"
)

// OrderRefund 订单退款成功消息处理
// Story 10.6: 售后退款触发的卡片处置策略
func OrderRefund(ctx context.Context, body []byte, cardMintService *digitalcardmint.Service, db *gorm.DB) error {
	var payload EventPayload
	if err := sonic.Unmarshal(body, &payload); err != nil {
		logc.Errorf(ctx, "OrderRefund 反序列化失败: %v", err)
		return err
	}

	ctx = payload.ToContext(ctx)
	LogWithEventPayload(ctx, "OrderRefund 收到订单退款成功事件, entityId=%d, action=%s", payload.EntityID, payload.Action)

	// 处理退款订单的卡片处置
	if err := processRefundedOrderCards(ctx, &payload, cardMintService, db); err != nil {
		LogWithEventPayload(ctx, "OrderRefund 卡片处置失败, orderId=%d, err=%v", payload.EntityID, err)
		return err
	}
	LogWithEventPayload(ctx, "OrderRefund 卡片处置完成, orderId=%d", payload.EntityID)

	return nil
}

type refundedOrderItemRow struct {
	ID           int64  `gorm:"column:id"`
	OrderID      int64  `gorm:"column:order_id"`
	SkuID        int64  `gorm:"column:sku_id"`
	SkuName      string `gorm:"column:sku_name"`
	RefundStatus int32  `gorm:"column:refund_status"`
}

func processRefundedOrderCards(ctx context.Context, payload *EventPayload, cardMintService *digitalcardmint.Service, db *gorm.DB) error {
	if payload == nil || payload.EntityID <= 0 || db == nil || cardMintService == nil {
		return nil
	}

	// 查询退款订单的订单明细
	var items []refundedOrderItemRow
	if err := db.WithContext(ctx).
		Table("oms_order_item").
		Select("id, order_id, sku_id, sku_name, refund_status").
		Where("order_id = ? AND is_deleted = 0", payload.EntityID).
		Find(&items).Error; err != nil {
		return err
	}
	if len(items) == 0 {
		return nil
	}

	// 获取订单的退款处置策略
	type orderRefundPolicyRow struct {
		RefundPolicy string `gorm:"column:refund_policy"`
	}

	// 查询每个订单明细对应的卡片资产的退款处置策略
	for _, item := range items {
		var policyRow orderRefundPolicyRow
		err := db.WithContext(ctx).
			Table("sms_card_instance").
			Select("refund_policy").
			Where("source_type = ? AND source_id = ? AND is_deleted = 0", "purchase", item.ID).
			Take(&policyRow).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				// 没有对应的卡片资产，跳过
				continue
			}
			logc.Errorf(ctx, "查询卡片资产退款策略失败, orderItemId=%d, err=%v", item.ID, err)
			continue
		}

		if policyRow.RefundPolicy == "" {
			// 没有配置退款策略，跳过
			continue
		}

		// 执行卡片处置
		result, err := cardMintService.HandleRefundCard处置(ctx, digitalcardmint.RefundCard处置Input{
			OrderID:      payload.EntityID,
			OrderItemID:  item.ID,
			RefundPolicy: policyRow.RefundPolicy,
			OperatorID:   payload.ActorID,
			Reason:       fmt.Sprintf("订单退款触发卡片处置, orderId=%d, orderItemId=%d", payload.EntityID, item.ID),
			TraceID:      payload.TraceID,
		})
		if err != nil {
			logc.Errorf(ctx, "卡片处置失败, orderItemId=%d, err=%v", item.ID, err)
			continue
		}

		if result != nil {
			LogWithEventPayload(ctx, "卡片处置成功, orderItemId=%d, assetInstanceId=%d, assetNo=%s, complianceStatus=%s, action=%s",
				item.ID, result.AssetInstanceID, result.AssetNo, result.ComplianceStatus, result.RefundAction)
		}
	}

	return nil
}
