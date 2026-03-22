package common

import (
	"context"

	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	"gorm.io/gorm"
)

type userScopeRow struct {
	PlatformID int64 `gorm:"column:platform_id"`
	TenantID   int64 `gorm:"column:tenant_id"`
	MerchantID int64 `gorm:"column:merchant_id"`
}

func ResolveActorScopeByUserName(ctx context.Context, db *gorm.DB, userName string) (pkgscope.GovernanceScope, error) {
	if userName == "" {
		return pkgscope.NormalizeGovernanceScope(pkgscope.SubjectTypePlatform, pkgscope.DefaultPlatformID, 0, 0)
	}

	var row userScopeRow
	if err := db.WithContext(ctx).
		Table("sys_user").
		Select("platform_id, tenant_id, merchant_id").
		Where("user_name = ?", userName).
		Take(&row).Error; err != nil {
		return pkgscope.GovernanceScope{}, err
	}

	return pkgscope.NormalizeGovernanceScope("", row.PlatformID, row.TenantID, row.MerchantID)
}

func ApplySubjectScope(ctx context.Context, db *gorm.DB, subjectID int64, current pkgscope.GovernanceScope) error {
	return db.WithContext(ctx).
		Table("cms_subject").
		Where("id = ?", subjectID).
		Updates(map[string]interface{}{
			"platform_id": current.PlatformID,
			"tenant_id":   current.TenantID,
			"merchant_id": current.MerchantID,
		}).Error
}
