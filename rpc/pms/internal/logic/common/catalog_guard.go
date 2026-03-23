package common

import (
	"context"
	"errors"

	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	"gorm.io/gorm"
)

type CatalogReferenceCheck struct {
	Table      string
	Column     string
	Message    string
	ScopeAware bool
}

type CatalogAttributeBindingRow struct {
	ID         int64 `gorm:"column:id"`
	PlatformID int64 `gorm:"column:platform_id"`
	TenantID   int64 `gorm:"column:tenant_id"`
	MerchantID int64 `gorm:"column:merchant_id"`
	Status     int32 `gorm:"column:status"`
}

func EnsureCatalogDeleteAllowed(ctx context.Context, db *gorm.DB, current pkgscope.GovernanceScope, ids []int64, checks []CatalogReferenceCheck) error {
	uniqueIDs := pkgscope.UniquePositiveIDs(ids)
	if len(uniqueIDs) == 0 {
		return errors.New("缺少有效资源ID")
	}

	for _, check := range checks {
		query := db.WithContext(ctx).Table(check.Table).Where(check.Column+" IN ?", uniqueIDs).Where("is_deleted = 0")
		if check.ScopeAware {
			query = query.Where("platform_id = ? AND tenant_id = ? AND merchant_id = ?", current.PlatformID, current.TenantID, current.MerchantID)
		}

		var count int64
		if err := query.Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return errors.New(check.Message)
		}
	}

	return nil
}

func ValidateCategoryAttributeBindings(ctx context.Context, db *gorm.DB, current pkgscope.GovernanceScope, attrIDs []int64) ([]int64, error) {
	uniqueIDs := pkgscope.UniquePositiveIDs(attrIDs)
	if len(uniqueIDs) == 0 {
		return uniqueIDs, nil
	}

	var rows []CatalogAttributeBindingRow
	if err := db.WithContext(ctx).
		Table("pms_product_attribute").
		Select("id, platform_id, tenant_id, merchant_id, status").
		Where("id IN ? AND is_deleted = 0", uniqueIDs).
		Find(&rows).Error; err != nil {
		return nil, err
	}

	if missing := pkgscope.MissingResourceIDs(toResourceScopeRows(rows), uniqueIDs); len(missing) > 0 {
		return nil, errors.New("绑定的商品属性不存在")
	}

	for _, row := range rows {
		rowScope := pkgscope.DefaultScope(row.PlatformID, row.TenantID, row.MerchantID)
		if !current.SameScope(rowScope) {
			return nil, errors.New("绑定的商品属性与当前分类作用域不一致")
		}
		if row.Status != 1 {
			return nil, errors.New("只能绑定启用状态的商品属性")
		}
	}

	return uniqueIDs, nil
}

func toResourceScopeRows(rows []CatalogAttributeBindingRow) []pkgscope.ResourceScopeRow {
	items := make([]pkgscope.ResourceScopeRow, 0, len(rows))
	for _, row := range rows {
		items = append(items, pkgscope.ResourceScopeRow{
			ID:         row.ID,
			PlatformID: row.PlatformID,
			TenantID:   row.TenantID,
			MerchantID: row.MerchantID,
		})
	}
	return items
}
