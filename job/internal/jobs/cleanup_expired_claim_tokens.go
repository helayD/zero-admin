package jobs

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/logc"
	"gorm.io/gorm"
)

type claimTokenRow struct {
	ID           int64      `gorm:"column:id"`
	Token        string     `gorm:"column:token"`
	Status       string     `gorm:"column:status"`
	ExpireAt     *time.Time `gorm:"column:expire_at"`
	ClaimedCount int32      `gorm:"column:claimed_count"`
	MaxClaims    int32      `gorm:"column:max_claims"`
	UpdatedAt    *time.Time `gorm:"column:updated_at"`
}

func (claimTokenRow) TableName() string {
	return "sms_card_claim_token"
}

// CleanupExpiredClaimTokens 扫描并清理已过期的分享凭证
// 将 status = active 且 expire_at < now() 的记录更新为 expired
func CleanupExpiredClaimTokens(ctx context.Context, db *gorm.DB) {
	if db == nil {
		logc.Errorf(ctx, "数据库未初始化，跳过过期凭证清理")
		return
	}

	now := time.Now()
	result := db.WithContext(ctx).
		Table(claimTokenRow{}.TableName()).
		Where("status = ? AND expire_at < ? AND is_deleted = 0", "active", now).
		Updates(map[string]interface{}{
			"status":     "expired",
			"updated_at": now,
		})

	if result.Error != nil {
		logc.Errorf(ctx, "清理过期分享凭证失败: %v", result.Error)
		return
	}

	logc.Infof(ctx, "过期分享凭证清理完成，更新数量: %d", result.RowsAffected)
}
