package common

import (
	"context"
	"errors"
	"fmt"
	"strings"

	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	"gorm.io/gorm"
)

const (
	ProductPublishStatusOffShelf int32 = 0
	ProductPublishStatusOnShelf  int32 = 1

	ProductVerifyStatusPending  int32 = 0
	ProductVerifyStatusApproved int32 = 1
	ProductVerifyStatusRejected int32 = 2

	ProductRecommendStatusOff int32 = 0
	ProductRecommendStatusOn  int32 = 1

	tenantStatusEnabled           int32 = 1
	merchantBusinessStatusEnabled int32 = 1
)

type ProductVisibilityRow struct {
	ID              int64  `gorm:"column:id"`
	Name            string `gorm:"column:name"`
	CategoryID      int64  `gorm:"column:category_id"`
	BrandID         int64  `gorm:"column:brand_id"`
	MainPic         string `gorm:"column:main_pic"`
	PublishStatus   int32  `gorm:"column:publish_status"`
	VerifyStatus    int32  `gorm:"column:verify_status"`
	RecommendStatus int32  `gorm:"column:recommend_status"`
	PreviewStatus   int32  `gorm:"column:preview_status"`
	Stock           int32  `gorm:"column:stock"`
	PlatformID      int64  `gorm:"column:platform_id"`
	TenantID        int64  `gorm:"column:tenant_id"`
	MerchantID      int64  `gorm:"column:merchant_id"`
}

type tenantStatusRow struct {
	Status int32 `gorm:"column:status"`
}

type merchantStatusRow struct {
	BusinessStatus int32 `gorm:"column:business_status"`
}

func EnsureProductsReviewReady(ctx context.Context, db *gorm.DB, current pkgscope.GovernanceScope, ids []int64) error {
	rows, err := LoadProductVisibilityRows(ctx, db, current, ids)
	if err != nil {
		return err
	}

	for _, row := range rows {
		if err := ensureProductDraftReady(ctx, db, row); err != nil {
			return fmt.Errorf("商品[%d]未满足送审条件: %w", row.ID, err)
		}
	}

	return nil
}

func EnsureProductsOwnerActive(ctx context.Context, db *gorm.DB, current pkgscope.GovernanceScope, ids []int64) error {
	rows, err := LoadProductVisibilityRows(ctx, db, current, ids)
	if err != nil {
		return err
	}

	for _, row := range rows {
		if err := ensureProductOwnerActive(ctx, db, row); err != nil {
			return fmt.Errorf("商品[%d]%w", row.ID, err)
		}
	}

	return nil
}

func EnsureProductsPublishable(ctx context.Context, db *gorm.DB, current pkgscope.GovernanceScope, ids []int64) error {
	rows, err := LoadProductVisibilityRows(ctx, db, current, ids)
	if err != nil {
		return err
	}

	for _, row := range rows {
		if err := ensureProductPublishable(ctx, db, row); err != nil {
			return fmt.Errorf("商品[%d]不能上架: %w", row.ID, err)
		}
	}

	return nil
}

func EnsureProductsRecommendable(ctx context.Context, db *gorm.DB, current pkgscope.GovernanceScope, ids []int64) error {
	rows, err := LoadProductVisibilityRows(ctx, db, current, ids)
	if err != nil {
		return err
	}

	for _, row := range rows {
		if err := ensureProductRecommendable(ctx, db, row); err != nil {
			return fmt.Errorf("商品[%d]不能推荐: %w", row.ID, err)
		}
	}

	return nil
}

func LoadProductVisibilityRows(ctx context.Context, db *gorm.DB, current pkgscope.GovernanceScope, ids []int64) ([]ProductVisibilityRow, error) {
	uniqueIDs := pkgscope.UniquePositiveIDs(ids)
	if len(uniqueIDs) == 0 {
		return nil, errors.New("缺少有效商品ID")
	}

	rows := make([]ProductVisibilityRow, 0, len(uniqueIDs))
	if err := pkgscope.ApplyGovernanceScope(
		db.WithContext(ctx).Table("pms_product_spu"),
		current,
		"",
	).Select(
		"id, name, category_id, brand_id, main_pic, publish_status, verify_status, recommend_status, preview_status, stock, platform_id, tenant_id, merchant_id",
	).Where("id IN ? AND is_deleted = 0", uniqueIDs).Find(&rows).Error; err != nil {
		return nil, err
	}

	if len(rows) != len(uniqueIDs) {
		return nil, errors.New("存在商品不存在或不在当前作用域内")
	}

	return rows, nil
}

func PartitionProductIndexIDs(ctx context.Context, db *gorm.DB, current pkgscope.GovernanceScope, ids []int64) ([]int64, []int64, error) {
	rows, err := LoadProductVisibilityRows(ctx, db, current, ids)
	if err != nil {
		return nil, nil, err
	}

	syncIDs := make([]int64, 0, len(rows))
	deleteIDs := make([]int64, 0, len(rows))
	for _, row := range rows {
		visible, err := shouldKeepProductIndexed(ctx, db, row)
		if err != nil {
			return nil, nil, err
		}
		if visible {
			syncIDs = append(syncIDs, row.ID)
			continue
		}
		deleteIDs = append(deleteIDs, row.ID)
	}

	return syncIDs, deleteIDs, nil
}

func ensureProductDraftReady(ctx context.Context, db *gorm.DB, row ProductVisibilityRow) error {
	ownerScope := pkgscope.DefaultScope(row.PlatformID, row.TenantID, row.MerchantID)

	if strings.TrimSpace(row.MainPic) == "" {
		return errors.New("缺少商品主图")
	}
	if row.Stock <= 0 {
		return errors.New("缺少有效库存")
	}
	if row.PreviewStatus != 0 {
		return errors.New("预告商品不可直接进入审核曝光链路")
	}
	if err := ensureProductOwnerActive(ctx, db, row); err != nil {
		return err
	}
	if err := EnsureScopedCategoryExists(ctx, db, ownerScope, row.CategoryID, "商品分类已失效或不在当前主体作用域"); err != nil {
		return err
	}
	if err := EnsureScopedBrandExists(ctx, db, ownerScope, row.BrandID, "商品品牌已失效或不在当前主体作用域"); err != nil {
		return err
	}
	if err := ensureProductHasSaleableSKU(ctx, db, row); err != nil {
		return err
	}

	return nil
}

func ensureProductPublishable(ctx context.Context, db *gorm.DB, row ProductVisibilityRow) error {
	if err := ensureProductDraftReady(ctx, db, row); err != nil {
		return err
	}
	if row.VerifyStatus != ProductVerifyStatusApproved {
		return errors.New("商品审核未通过")
	}

	return nil
}

func ensureProductRecommendable(ctx context.Context, db *gorm.DB, row ProductVisibilityRow) error {
	if err := ensureProductPublishable(ctx, db, row); err != nil {
		return err
	}
	if row.PublishStatus != ProductPublishStatusOnShelf {
		return errors.New("商品未上架")
	}

	return nil
}

func ensureProductHasSaleableSKU(ctx context.Context, db *gorm.DB, row ProductVisibilityRow) error {
	var count int64
	if err := db.WithContext(ctx).
		Table("pms_product_sku").
		Where("spu_id = ? AND is_deleted = 0 AND stock > 0", row.ID).
		Where("platform_id = ? AND tenant_id = ? AND merchant_id = ?", row.PlatformID, row.TenantID, row.MerchantID).
		Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return errors.New("缺少可售SKU")
	}

	return nil
}

func ensureProductOwnerActive(ctx context.Context, db *gorm.DB, row ProductVisibilityRow) error {
	if row.TenantID > 0 {
		var tenant tenantStatusRow
		if err := db.WithContext(ctx).Table("sys_tenant").Select("status").Where("id = ?", row.TenantID).Take(&tenant).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("商品所属租户不存在")
			}
			return err
		}
		if tenant.Status != tenantStatusEnabled {
			return errors.New("商品所属租户已停用")
		}
	}

	if row.MerchantID > 0 {
		var merchant merchantStatusRow
		if err := db.WithContext(ctx).Table("sys_merchant").Select("business_status").Where("id = ?", row.MerchantID).Take(&merchant).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.New("商品所属商户不存在")
			}
			return err
		}
		if merchant.BusinessStatus != merchantBusinessStatusEnabled {
			return errors.New("商品所属商户已停用")
		}
	}

	return nil
}

func shouldKeepProductIndexed(ctx context.Context, db *gorm.DB, row ProductVisibilityRow) (bool, error) {
	if row.VerifyStatus != ProductVerifyStatusApproved {
		return false, nil
	}
	if row.PublishStatus != ProductPublishStatusOnShelf {
		return false, nil
	}
	if row.PreviewStatus != 0 {
		return false, nil
	}
	if err := ensureProductOwnerActive(ctx, db, row); err != nil {
		if isSoftVisibilityBlock(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func isSoftVisibilityBlock(err error) bool {
	if err == nil {
		return false
	}
	message := err.Error()
	return strings.Contains(message, "租户不存在") ||
		strings.Contains(message, "租户已停用") ||
		strings.Contains(message, "商户不存在") ||
		strings.Contains(message, "商户已停用")
}
