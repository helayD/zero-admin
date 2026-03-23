package common

import (
	"context"
	"errors"
	"fmt"

	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	"gorm.io/gorm"
)

type catalogReferenceCheck struct {
	table   string
	where   string
	args    []interface{}
	message string
}

func ensureNoCatalogReferences(ctx context.Context, db *gorm.DB, checks ...catalogReferenceCheck) error {
	for _, check := range checks {
		var count int64
		if err := db.WithContext(ctx).Table(check.table).Where(check.where, check.args...).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return errors.New(check.message)
		}
	}
	return nil
}

func EnsureBrandDeleteAllowed(ctx context.Context, db *gorm.DB, ids []int64) error {
	uniqueIDs := pkgscope.UniquePositiveIDs(ids)
	if len(uniqueIDs) == 0 {
		return errors.New("缺少有效商品品牌ID")
	}

	return ensureNoCatalogReferences(ctx, db,
		catalogReferenceCheck{
			table:   "pms_product_spu",
			where:   "brand_id IN ? AND is_deleted = 0",
			args:    []interface{}{uniqueIDs},
			message: "商品品牌已被商品SPU引用，暂不可删除",
		},
	)
}

func EnsureCategoryDeleteAllowed(ctx context.Context, db *gorm.DB, ids []int64) error {
	uniqueIDs := pkgscope.UniquePositiveIDs(ids)
	if len(uniqueIDs) == 0 {
		return errors.New("缺少有效商品分类ID")
	}

	return ensureNoCatalogReferences(ctx, db,
		catalogReferenceCheck{
			table:   "pms_product_category",
			where:   "parent_id IN ? AND is_deleted = 0",
			args:    []interface{}{uniqueIDs},
			message: "商品分类下仍存在子分类，暂不可删除",
		},
		catalogReferenceCheck{
			table:   "pms_product_attribute_group",
			where:   "category_id IN ? AND is_deleted = 0",
			args:    []interface{}{uniqueIDs},
			message: "商品分类已被属性分组引用，暂不可删除",
		},
		catalogReferenceCheck{
			table:   "pms_product_spec",
			where:   "category_id IN ? AND is_deleted = 0",
			args:    []interface{}{uniqueIDs},
			message: "商品分类已被商品规格引用，暂不可删除",
		},
		catalogReferenceCheck{
			table:   "pms_product_category_attribute_relation",
			where:   "product_category_id IN ?",
			args:    []interface{}{uniqueIDs},
			message: "商品分类仍绑定商品属性关系，暂不可删除",
		},
		catalogReferenceCheck{
			table:   "pms_product_spu",
			where:   "category_id IN ? AND is_deleted = 0",
			args:    []interface{}{uniqueIDs},
			message: "商品分类已被商品SPU引用，暂不可删除",
		},
	)
}

func EnsureAttributeDeleteAllowed(ctx context.Context, db *gorm.DB, ids []int64) error {
	uniqueIDs := pkgscope.UniquePositiveIDs(ids)
	if len(uniqueIDs) == 0 {
		return errors.New("缺少有效商品属性ID")
	}

	return ensureNoCatalogReferences(ctx, db,
		catalogReferenceCheck{
			table:   "pms_product_category_attribute_relation",
			where:   "product_attribute_id IN ?",
			args:    []interface{}{uniqueIDs},
			message: "商品属性已绑定商品分类，暂不可删除",
		},
		catalogReferenceCheck{
			table:   "pms_product_attribute_value",
			where:   "attribute_id IN ? AND is_deleted = 0",
			args:    []interface{}{uniqueIDs},
			message: "商品属性已被商品属性值引用，暂不可删除",
		},
	)
}

func EnsureAttributeGroupDeleteAllowed(ctx context.Context, db *gorm.DB, ids []int64) error {
	uniqueIDs := pkgscope.UniquePositiveIDs(ids)
	if len(uniqueIDs) == 0 {
		return errors.New("缺少有效商品属性分组ID")
	}

	return ensureNoCatalogReferences(ctx, db,
		catalogReferenceCheck{
			table:   "pms_product_attribute",
			where:   "group_id IN ? AND is_deleted = 0",
			args:    []interface{}{uniqueIDs},
			message: "商品属性分组下仍存在商品属性，暂不可删除",
		},
	)
}

func EnsureSpecDeleteAllowed(ctx context.Context, db *gorm.DB, ids []int64) error {
	uniqueIDs := pkgscope.UniquePositiveIDs(ids)
	if len(uniqueIDs) == 0 {
		return errors.New("缺少有效商品规格ID")
	}

	return ensureNoCatalogReferences(ctx, db,
		catalogReferenceCheck{
			table:   "pms_product_spec_value",
			where:   "spec_id IN ? AND is_deleted = 0",
			args:    []interface{}{uniqueIDs},
			message: "商品规格下仍存在规格值，暂不可删除",
		},
	)
}

type scopedCatalogAttributeRow struct {
	ID         int64
	Status     int32
	PlatformID int64
	TenantID   int64
	MerchantID int64
}

func EnsureCategoryAttributeBindings(ctx context.Context, db *gorm.DB, current pkgscope.GovernanceScope, attrIDs []int64) error {
	uniqueIDs := pkgscope.UniquePositiveIDs(attrIDs)
	if len(uniqueIDs) == 0 {
		return nil
	}

	var rows []scopedCatalogAttributeRow
	if err := db.WithContext(ctx).
		Table("pms_product_attribute").
		Select("id, status, platform_id, tenant_id, merchant_id").
		Where("id IN ? AND is_deleted = 0", uniqueIDs).
		Scan(&rows).Error; err != nil {
		return err
	}

	rowMap := make(map[int64]scopedCatalogAttributeRow, len(rows))
	for _, row := range rows {
		rowMap[row.ID] = row
	}

	for _, attrID := range uniqueIDs {
		row, ok := rowMap[attrID]
		if !ok {
			return fmt.Errorf("绑定的商品属性不存在: %d", attrID)
		}

		if !current.SameScope(pkgscope.DefaultScope(row.PlatformID, row.TenantID, row.MerchantID)) {
			return fmt.Errorf("绑定的商品属性与当前主体作用域不一致: %d", attrID)
		}

		if row.Status != 1 {
			return fmt.Errorf("只能绑定启用状态的商品属性: %d", attrID)
		}
	}

	return nil
}
