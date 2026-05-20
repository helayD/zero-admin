package memberinfoservicelogic

import (
	"context"
	"time"

	"github.com/feihua/zero-admin/rpc/ums/gen/model"
	"gorm.io/gorm"
)

const (
	dailyLoginPointsGrantType = "daily_login_points"
	dailyLoginPoints          = 10
)

func grantDailyLoginPoints(ctx context.Context, db *gorm.DB, memberID int64, now time.Time) error {
	if db == nil || memberID <= 0 {
		return nil
	}
	grantKey := now.Format("2006-01-02")
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Exec(dailyLoginPointsInsertSQL(tx),
			memberID,
			dailyLoginPointsGrantType,
			grantKey,
			dailyLoginPoints,
			grantKey,
			"每日登录赠送积分",
			now,
		)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return nil
		}

		update := tx.Table("ums_member_info").
			Where("member_id = ? AND is_deleted = 0", memberID).
			Updates(map[string]interface{}{
				"points":       gorm.Expr("points + ?", dailyLoginPoints),
				"total_points": gorm.Expr("total_points + ?", dailyLoginPoints),
				"update_time":  now,
			})
		if update.Error != nil {
			return update.Error
		}
		if update.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}

		return tx.Create(&model.UmsMemberPointsLog{
			MemberID:     memberID,
			ChangeType:   1,
			ChangePoints: dailyLoginPoints,
			SourceType:   2,
			Description:  "每日登录赠送积分",
			OperateMan:   "system",
			OperateNote:  "daily_login:" + grantKey,
			CreateTime:   now,
		}).Error
	})
}

func dailyLoginPointsInsertSQL(db *gorm.DB) string {
	if db != nil && db.Dialector != nil && db.Dialector.Name() == "sqlite" {
		return `
INSERT OR IGNORE INTO ums_member_lottery_grant_log
    (member_id, grant_type, grant_key, grant_times, source_ref, description, create_time)
VALUES
    (?, ?, ?, ?, ?, ?, ?)`
	}
	return `
INSERT IGNORE INTO ums_member_lottery_grant_log
    (member_id, grant_type, grant_key, grant_times, source_ref, description, create_time)
VALUES
    (?, ?, ?, ?, ?, ?, ?)`
}
