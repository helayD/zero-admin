package jobs

import (
	"context"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/logc"
	"gorm.io/gorm"
)

type redemptionOrderRow struct {
	ID             int64      `gorm:"column:id"`
	OrderNo        string     `gorm:"column:order_no"`
	CardInstanceID int64      `gorm:"column:card_instance_id"`
	Status         string     `gorm:"column:status"`
	CreatedAt      *time.Time `gorm:"column:created_at"`
	UpdatedAt      *time.Time `gorm:"column:updated_at"`
}

func (redemptionOrderRow) TableName() string {
	return "sms_card_redemption_order"
}

type cardInstanceRow struct {
	ID          int64  `gorm:"column:id"`
	AssetStatus string `gorm:"column:asset_status"`
}

func (cardInstanceRow) TableName() string {
	return "sms_card_instance"
}

type cardAssetLogRow struct {
	AssetInstanceID       int64  `gorm:"column:asset_instance_id"`
	FromStatus            string `gorm:"column:from_status"`
	ToStatus              string `gorm:"column:to_status"`
	OperationType         string `gorm:"column:operation_type"`
	OperatorType          string `gorm:"column:operator_type"`
	TraceID               string `gorm:"column:trace_id"`
	ReasonCode            string `gorm:"column:reason_code"`
	ReasonText            string `gorm:"column:reason_text"`
	PayloadJSON           string `gorm:"column:payload_json"`
}

func (cardAssetLogRow) TableName() string {
	return "sms_card_asset_log"
}

// CheckRedemptionOrderTimeout 扫描超时的 pending 提货单并自动取消
// pending 状态超过 24 小时的提货单会被取消，并恢复卡片状态为 claimed
func CheckRedemptionOrderTimeout(ctx context.Context, db *gorm.DB) {
	if db == nil {
		logc.Errorf(ctx, "数据库未初始化，跳过提货单超时检查")
		return
	}

	cutoff := time.Now().Add(-24 * time.Hour)

	var orders []redemptionOrderRow
	if err := db.WithContext(ctx).
		Table(redemptionOrderRow{}.TableName()).
		Where("status = ? AND created_at < ? AND is_deleted = 0", "pending", cutoff).
		Find(&orders).Error; err != nil {
		logc.Errorf(ctx, "查询超时提货单失败: %v", err)
		return
	}

	if len(orders) == 0 {
		logc.Infof(ctx, "提货单超时检查完成，无超时订单")
		return
	}

	now := time.Now()
	for _, order := range orders {
		err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			// 更新提货单状态为 cancelled
			if err := tx.Table(redemptionOrderRow{}.TableName()).
				Where("id = ?", order.ID).
				Updates(map[string]interface{}{
					"status":       "cancelled",
					"cancel_reason": "超时自动取消",
					"updated_at":   now,
				}).Error; err != nil {
				return fmt.Errorf("更新提货单状态失败: %w", err)
			}

			// 恢复卡片状态为 claimed
			if err := tx.Table(cardInstanceRow{}.TableName()).
				Where("id = ?", order.CardInstanceID).
				Updates(map[string]interface{}{
					"asset_status": "claimed",
					"update_time":  now,
				}).Error; err != nil {
				return fmt.Errorf("恢复卡片状态失败: %w", err)
			}

			// 记录审计日志
			logRow := &cardAssetLogRow{
				AssetInstanceID: order.CardInstanceID,
				FromStatus:      "pending_redemption",
				ToStatus:        "claimed",
				OperationType:   "redemption_order_cancelled",
				OperatorType:    "system",
				TraceID:         fmt.Sprintf("job-timeout-%d", order.ID),
				ReasonCode:      "timeout",
				ReasonText:      "提货单超时自动取消，恢复卡片状态",
				PayloadJSON:     fmt.Sprintf(`{"orderId":%d,"orderNo":"%s","reason":"timeout_auto_cancel"}`, order.ID, order.OrderNo),
			}
			if err := tx.Table(logRow.TableName()).Create(logRow).Error; err != nil {
				return fmt.Errorf("记录审计日志失败: %w", err)
			}

			return nil
		})

		if err != nil {
			logc.Errorf(ctx, "处理超时提货单失败, orderId=%d: %v", order.ID, err)
			continue
		}
		logc.Infof(ctx, "超时提货单已自动取消, orderId=%d, orderNo=%s", order.ID, order.OrderNo)
	}

	logc.Infof(ctx, "提货单超时检查完成，处理数量: %d", len(orders))
}
