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
	ID                int64  `gorm:"column:id"`
	Name              string `gorm:"column:name"`
	CategoryID        int64  `gorm:"column:category_id"`
	BrandID           int64  `gorm:"column:brand_id"`
	MainPic           string `gorm:"column:main_pic"`
	PublishStatus     int32  `gorm:"column:publish_status"`
	VerifyStatus      int32  `gorm:"column:verify_status"`
	RecommendStatus   int32  `gorm:"column:recommend_status"`
	PreviewStatus     int32  `gorm:"column:preview_status"`
	Stock             int32  `gorm:"column:stock"`
	FulfillmentMode   string `gorm:"column:fulfillment_mode"`
	FulfillmentRuleID int64  `gorm:"column:fulfillment_rule_id"`
	PlatformID        int64  `gorm:"column:platform_id"`
	TenantID          int64  `gorm:"column:tenant_id"`
	MerchantID        int64  `gorm:"column:merchant_id"`
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
		"id, name, category_id, brand_id, main_pic, publish_status, verify_status, recommend_status, preview_status, stock, fulfillment_mode, fulfillment_rule_id, platform_id, tenant_id, merchant_id",
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
	// 履约模式校验（Story 10.6）- 使用结构化错误
	errs := ValidateFulfillmentModeWithDetails(ctx, db, row)
	if errs.HasErrors() {
		return errs
	}

	return nil
}

// validateFulfillmentMode 校验商品履约模式配置的有效性
func validateFulfillmentMode(ctx context.Context, db *gorm.DB, row ProductVisibilityRow) error {
	// 1. 校验履约模式取值
	if row.FulfillmentMode == "" {
		// 兼容旧数据：未配置履约模式的商品默认为实物发货
		return nil
	}
	if row.FulfillmentMode != "physical_delivery" && row.FulfillmentMode != "digital_asset" {
		return fmt.Errorf("履约模式取值非法[%s]，仅支持 physical_delivery 或 digital_asset", row.FulfillmentMode)
	}

	// 2. 提货卡模式的额外校验
	if row.FulfillmentMode == "digital_asset" {
		if row.FulfillmentRuleID <= 0 {
			return errors.New("提货卡模式商品必须绑定有效发卡规则")
		}
		// 校验发卡规则是否存在且有效
		if err := validateFulfillmentRule(ctx, db, row); err != nil {
			return err
		}
	}

	return nil
}

// fulfillmentRuleRow 发卡规则查询结果
type fulfillmentRuleRow struct {
	ID             int64  `gorm:"column:id"`
	RuleStatus     int32  `gorm:"column:rule_status"`
	CardTemplateID int64  `gorm:"column:card_template_id"`
	PlatformID     int64  `gorm:"column:platform_id"`
	TenantID       int64  `gorm:"column:tenant_id"`
	MerchantID     int64  `gorm:"column:merchant_id"`
}

// cardTemplateRow 卡片模板查询结果
type cardTemplateRow struct {
	ID     int64 `gorm:"column:id"`
	Status int32 `gorm:"column:status"`
}

// validateFulfillmentRule 校验发卡规则的有效性
func validateFulfillmentRule(ctx context.Context, db *gorm.DB, row ProductVisibilityRow) error {
	// 查询发卡规则
	var rule fulfillmentRuleRow
	err := db.WithContext(ctx).
		Table("sms_product_fulfillment_rule").
		Select("id, rule_status, card_template_id, platform_id, tenant_id, merchant_id").
		Where("id = ? AND is_deleted = 0", row.FulfillmentRuleID).
		Take(&rule).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("发卡规则[%d]不存在或已删除", row.FulfillmentRuleID)
		}
		return fmt.Errorf("查询发卡规则失败: %w", err)
	}

	// 校验发卡规则状态
	if rule.RuleStatus != 1 {
		return fmt.Errorf("发卡规则[%d]已禁用，请先启用规则或更换其他规则", row.FulfillmentRuleID)
	}

	// 校验作用域一致性
	if rule.PlatformID != row.PlatformID || rule.TenantID != row.TenantID || rule.MerchantID != row.MerchantID {
		return fmt.Errorf("发卡规则[%d]的作用域与商品不一致，规则归属(platform:%d,tenant:%d,merchant:%d)，商品归属(platform:%d,tenant:%d,merchant:%d)",
			row.FulfillmentRuleID, rule.PlatformID, rule.TenantID, rule.MerchantID,
			row.PlatformID, row.TenantID, row.MerchantID)
	}

	// 校验关联的卡片模板是否存在且有效
	if rule.CardTemplateID > 0 {
		var template cardTemplateRow
		err := db.WithContext(ctx).
			Table("sms_card_template").
			Select("id, status").
			Where("id = ? AND is_deleted = 0", rule.CardTemplateID).
			Take(&template).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("发卡规则[%d]关联的卡片模板[%d]不存在或已删除", row.FulfillmentRuleID, rule.CardTemplateID)
			}
			return fmt.Errorf("查询卡片模板失败: %w", err)
		}
		// 注：status 校验取决于卡片模板的状态定义，这里假设 0-禁用 1-启用
		// 如果模板状态定义不同，需要调整
		if template.Status == 0 {
			return fmt.Errorf("发卡规则[%d]关联的卡片模板[%d]已禁用，请先启用模板或更换规则", row.FulfillmentRuleID, rule.CardTemplateID)
		}
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
