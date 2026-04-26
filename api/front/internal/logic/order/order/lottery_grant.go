package order

import (
	"context"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/logc"
	"gorm.io/gorm"
)

const (
	orderCompletedLotteryGrantType  = "order_completed"
	orderCompletedLotteryGrantTimes = 1
)

func grantOrderCompletedLotteryTimes(ctx context.Context, db *gorm.DB, memberID int64, orderID int64) {
	if db == nil || memberID <= 0 || orderID <= 0 {
		return
	}
	now := time.Now()
	grantKey := fmt.Sprintf("%d", orderID)
	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Exec(`
INSERT IGNORE INTO ums_member_lottery_grant_log
    (member_id, grant_type, grant_key, grant_times, source_ref, description, create_time)
VALUES
    (?, ?, ?, ?, ?, ?, ?)`,
			memberID,
			orderCompletedLotteryGrantType,
			grantKey,
			orderCompletedLotteryGrantTimes,
			grantKey,
			"订单完成赠送抽卡次数",
			now,
		)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return nil
		}
		return tx.Table("ums_member_info").
			Where("member_id = ? AND is_deleted = 0", memberID).
			Updates(map[string]interface{}{
				"lottery_times": gorm.Expr("lottery_times + ?", orderCompletedLotteryGrantTimes),
				"update_time":   now,
			}).Error
	})
	if err != nil {
		logc.Errorf(ctx, "订单完成赠送抽卡次数失败,memberId:%d,orderId:%d,异常:%s", memberID, orderID, err.Error())
	}
}
