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

func ResolveActorScope(ctx context.Context, db *gorm.DB, actorID int64) (pkgscope.GovernanceScope, error) {
	if actorID <= 0 {
		return pkgscope.NormalizeGovernanceScope(pkgscope.SubjectTypePlatform, pkgscope.DefaultPlatformID, 0, 0)
	}

	var row userScopeRow
	if err := db.WithContext(ctx).
		Table("sys_user").
		Select("platform_id, tenant_id, merchant_id").
		Where("id = ?", actorID).
		Take(&row).Error; err != nil {
		return pkgscope.GovernanceScope{}, err
	}

	return pkgscope.NormalizeGovernanceScope("", row.PlatformID, row.TenantID, row.MerchantID)
}

func ApplyProductScope(ctx context.Context, db *gorm.DB, productID int64, current pkgscope.GovernanceScope) error {
	scopeValues := map[string]interface{}{
		"platform_id": current.PlatformID,
		"tenant_id":   current.TenantID,
		"merchant_id": current.MerchantID,
	}

	tx := db.WithContext(ctx)
	if err := tx.Table("pms_product_spu").Where("id = ?", productID).Updates(scopeValues).Error; err != nil {
		return err
	}

	return tx.Table("pms_product_sku").Where("spu_id = ?", productID).Updates(scopeValues).Error
}
