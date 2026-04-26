package memberinfoservicelogic

import (
	"context"
	"time"

	"gorm.io/gorm"
)

const (
	dailyLoginLotteryGrantType  = "daily_login"
	dailyLoginLotteryGrantTimes = 3
)

func grantDailyLoginLotteryTimes(ctx context.Context, db *gorm.DB, memberID int64, now time.Time) error {
	if db == nil || memberID <= 0 {
		return nil
	}
	grantKey := now.Format("2006-01-02")
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Exec(`
INSERT IGNORE INTO ums_member_lottery_grant_log
    (member_id, grant_type, grant_key, grant_times, source_ref, description, create_time)
VALUES
    (?, ?, ?, ?, ?, ?, ?)`,
			memberID,
			dailyLoginLotteryGrantType,
			grantKey,
			dailyLoginLotteryGrantTimes,
			grantKey,
			"每日登录赠送抽卡次数",
			now,
		)
		if result.Error != nil {
			return result.Error
		}

		updates := map[string]interface{}{
			"last_login":  now,
			"update_time": now,
		}
		if result.RowsAffected > 0 {
			updates["lottery_times"] = gorm.Expr("lottery_times + ?", dailyLoginLotteryGrantTimes)
		}

		return tx.Table("ums_member_info").
			Where("member_id = ? AND is_deleted = 0", memberID).
			Updates(updates).Error
	})
}
