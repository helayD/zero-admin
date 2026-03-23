package common

import (
	"context"
	"errors"

	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	"gorm.io/gorm"
)

func EnsureAttributeScope(ctx context.Context, db *gorm.DB, current pkgscope.GovernanceScope, ids []int64, action string, operatorID int64, operatorName, requestSummary string) ([]pkgscope.ResourceScopeRow, error) {
	return ensureCatalogWriteScope(ctx, db, current, "pms_product_attribute", ids, action, "product_attribute", operatorID, operatorName, requestSummary, "当前主体无权修改所选商品属性", "商品属性不存在")
}

func EnsureAttributeGroupScope(ctx context.Context, db *gorm.DB, current pkgscope.GovernanceScope, ids []int64, action string, operatorID int64, operatorName, requestSummary string) ([]pkgscope.ResourceScopeRow, error) {
	return ensureCatalogWriteScope(ctx, db, current, "pms_product_attribute_group", ids, action, "product_attribute_group", operatorID, operatorName, requestSummary, "当前主体无权修改所选商品属性分组", "商品属性分组不存在")
}

func EnsureSpecScope(ctx context.Context, db *gorm.DB, current pkgscope.GovernanceScope, ids []int64, action string, operatorID int64, operatorName, requestSummary string) ([]pkgscope.ResourceScopeRow, error) {
	return ensureCatalogWriteScope(ctx, db, current, "pms_product_spec", ids, action, "product_spec", operatorID, operatorName, requestSummary, "当前主体无权修改所选商品规格", "商品规格不存在")
}

func EnsureSpecValueScope(ctx context.Context, db *gorm.DB, current pkgscope.GovernanceScope, ids []int64, action string, operatorID int64, operatorName, requestSummary string) ([]pkgscope.ResourceScopeRow, error) {
	return ensureCatalogWriteScope(ctx, db, current, "pms_product_spec_value", ids, action, "product_spec_value", operatorID, operatorName, requestSummary, "当前主体无权修改所选商品规格值", "商品规格值不存在")
}

func EnsureScopedCategoryExists(ctx context.Context, db *gorm.DB, current pkgscope.GovernanceScope, categoryID int64, message string) error {
	if categoryID <= 0 { return errors.New("缺少有效商品分类ID") }
	var row pkgscope.ResourceScopeRow
	if err := db.WithContext(ctx).Table("pms_product_category").Select("id, platform_id, tenant_id, merchant_id").Where("id = ? AND is_deleted = 0", categoryID).Take(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { return errors.New("商品分类不存在") }
		return err
	}
	if !current.SameScope(pkgscope.DefaultScope(row.PlatformID, row.TenantID, row.MerchantID)) { return errors.New(message) }
	return nil
}

func EnsureScopedAttributeGroupExists(ctx context.Context, db *gorm.DB, current pkgscope.GovernanceScope, groupID int64, message string) error {
	if groupID <= 0 { return errors.New("缺少有效属性分组ID") }
	var row pkgscope.ResourceScopeRow
	if err := db.WithContext(ctx).Table("pms_product_attribute_group").Select("id, platform_id, tenant_id, merchant_id").Where("id = ? AND is_deleted = 0", groupID).Take(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { return errors.New("商品属性分组不存在") }
		return err
	}
	if !current.SameScope(pkgscope.DefaultScope(row.PlatformID, row.TenantID, row.MerchantID)) { return errors.New(message) }
	return nil
}

func EnsureScopedSpecExists(ctx context.Context, db *gorm.DB, current pkgscope.GovernanceScope, specID int64, message string) error {
	if specID <= 0 { return errors.New("缺少有效商品规格ID") }
	var row pkgscope.ResourceScopeRow
	if err := db.WithContext(ctx).Table("pms_product_spec").Select("id, platform_id, tenant_id, merchant_id").Where("id = ? AND is_deleted = 0", specID).Take(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) { return errors.New("商品规格不存在") }
		return err
	}
	if !current.SameScope(pkgscope.DefaultScope(row.PlatformID, row.TenantID, row.MerchantID)) { return errors.New(message) }
	return nil
}
