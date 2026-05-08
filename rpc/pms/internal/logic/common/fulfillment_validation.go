package common

import (
	"context"
	"errors"
	"fmt"
	"strings"

	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	"gorm.io/gorm"
)

// FulfillmentValidationError 履约模式校验错误
type FulfillmentValidationError struct {
	Code      string `json:"code"`       // 错误码
	Field     string `json:"field"`      // 冲突字段
	Message   string `json:"message"`    // 错误信息
	Suggestion string `json:"suggestion"` // 修复建议
}

func (e *FulfillmentValidationError) Error() string {
	return fmt.Sprintf("[%s] %s (字段: %s, 建议: %s)", e.Code, e.Message, e.Field, e.Suggestion)
}

// FulfillmentValidationErrors 多个校验错误
type FulfillmentValidationErrors struct {
	Errors []*FulfillmentValidationError `json:"errors"`
}

func (e *FulfillmentValidationErrors) Error() string {
	var messages []string
	for _, err := range e.Errors {
		messages = append(messages, err.Error())
	}
	return strings.Join(messages, "; ")
}

// HasErrors 是否有错误
func (e *FulfillmentValidationErrors) HasErrors() bool {
	return len(e.Errors) > 0
}

// AddError 添加错误
func (e *FulfillmentValidationErrors) AddError(code, field, message, suggestion string) {
	e.Errors = append(e.Errors, &FulfillmentValidationError{
		Code:       code,
		Field:      field,
		Message:    message,
		Suggestion: suggestion,
	})
}

// NewFulfillmentValidationErrors 创建校验错误集合
func NewFulfillmentValidationErrors() *FulfillmentValidationErrors {
	return &FulfillmentValidationErrors{
		Errors: make([]*FulfillmentValidationError, 0),
	}
}

// 预定义错误码
const (
	FulfillmentErrCodeInvalidMode       = "FULFILLMENT_INVALID_MODE"
	FulfillmentErrCodeMissingRule       = "FULFILLMENT_MISSING_RULE"
	FulfillmentErrCodeRuleNotFound      = "FULFILLMENT_RULE_NOT_FOUND"
	FulfillmentErrCodeRuleDisabled      = "FULFILLMENT_RULE_DISABLED"
	FulfillmentErrCodeScopeMismatch     = "FULFILLMENT_SCOPE_MISMATCH"
	FulfillmentErrCodeTemplateNotFound  = "FULFILLMENT_TEMPLATE_NOT_FOUND"
	FulfillmentErrCodeTemplateDisabled  = "FULFILLMENT_TEMPLATE_DISABLED"
	FulfillmentErrCodeConfigConflict    = "FULFILLMENT_CONFIG_CONFLICT"
)

// ValidateFulfillmentModeWithDetails 校验履约模式并返回详细错误
func ValidateFulfillmentModeWithDetails(ctx context.Context, db *gorm.DB, row ProductVisibilityRow) *FulfillmentValidationErrors {
	errs := NewFulfillmentValidationErrors()

	// 1. 校验履约模式取值
	if row.FulfillmentMode == "" {
		// 兼容旧数据：未配置履约模式的商品默认为实物发货
		return errs
	}
	if row.FulfillmentMode != "physical_delivery" && row.FulfillmentMode != "digital_asset" {
		errs.AddError(
			FulfillmentErrCodeInvalidMode,
			"fulfillment_mode",
			fmt.Sprintf("履约模式取值非法[%s]", row.FulfillmentMode),
			"请将履约模式设置为 physical_delivery（实物发货）或 digital_asset（提货卡）",
		)
		return errs
	}

	// 2. 提货卡模式的额外校验
	if row.FulfillmentMode == "digital_asset" {
		if row.FulfillmentRuleID <= 0 {
			errs.AddError(
				FulfillmentErrCodeMissingRule,
				"fulfillment_rule_id",
				"提货卡模式商品必须绑定发卡规则",
				"请在商品编辑页面选择一个有效的发卡规则",
			)
			return errs
		}

		// 校验发卡规则
		validateFulfillmentRuleWithDetails(ctx, db, row, errs)
	}

	return errs
}

// validateFulfillmentRuleWithDetails 校验发卡规则并返回详细错误
func validateFulfillmentRuleWithDetails(ctx context.Context, db *gorm.DB, row ProductVisibilityRow, errs *FulfillmentValidationErrors) {
	// 查询发卡规则
	var rule fulfillmentRuleRow
	err := db.WithContext(ctx).
		Table("sms_product_fulfillment_rule").
		Select("id, rule_status, card_template_id, platform_id, tenant_id, merchant_id").
		Where("id = ? AND is_deleted = 0", row.FulfillmentRuleID).
		Take(&rule).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			errs.AddError(
				FulfillmentErrCodeRuleNotFound,
				"fulfillment_rule_id",
				fmt.Sprintf("发卡规则[%d]不存在或已删除", row.FulfillmentRuleID),
				"请重新选择一个有效的发卡规则",
			)
			return
		}
		errs.AddError(
			FulfillmentErrCodeRuleNotFound,
			"fulfillment_rule_id",
			fmt.Sprintf("查询发卡规则失败: %v", err),
			"请稍后重试或联系管理员",
		)
		return
	}

	// 校验发卡规则状态
	if rule.RuleStatus != 1 {
		errs.AddError(
			FulfillmentErrCodeRuleDisabled,
			"fulfillment_rule_id",
			fmt.Sprintf("发卡规则[%d]已禁用", row.FulfillmentRuleID),
			"请启用该发卡规则，或更换其他已启用的规则",
		)
	}

	// 校验作用域一致性
	if rule.PlatformID != row.PlatformID || rule.TenantID != row.TenantID || rule.MerchantID != row.MerchantID {
		errs.AddError(
			FulfillmentErrCodeScopeMismatch,
			"fulfillment_rule_id",
			fmt.Sprintf("发卡规则[%d]的作用域与商品不一致", row.FulfillmentRuleID),
			fmt.Sprintf("请更换为与商品相同作用域(platform:%d,tenant:%d,merchant:%d)的发卡规则", row.PlatformID, row.TenantID, row.MerchantID),
		)
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
				errs.AddError(
					FulfillmentErrCodeTemplateNotFound,
					"card_template_id",
					fmt.Sprintf("发卡规则[%d]关联的卡片模板[%d]不存在或已删除", row.FulfillmentRuleID, rule.CardTemplateID),
					"请修改发卡规则，关联一个有效的卡片模板",
				)
			}
			return
		}
		// 注：status 校验取决于卡片模板的状态定义
		if template.Status == 0 {
			errs.AddError(
				FulfillmentErrCodeTemplateDisabled,
				"card_template_id",
				fmt.Sprintf("发卡规则[%d]关联的卡片模板[%d]已禁用", row.FulfillmentRuleID, rule.CardTemplateID),
				"请启用卡片模板，或更换关联其他已启用的模板",
			)
		}
	}
}

// CheckProductFulfillmentConflict 检查已上架商品的履约模式配置冲突
// 用于 Task 3.5: 展示"配置冲突"或"生效异常"状态
func CheckProductFulfillmentConflict(ctx context.Context, db *gorm.DB, row ProductVisibilityRow) (bool, string) {
	// 只检查已上架的商品
	if row.PublishStatus != ProductPublishStatusOnShelf {
		return false, ""
	}

	// 只检查提货卡模式
	if row.FulfillmentMode != "digital_asset" {
		return false, ""
	}

	// 检查发卡规则是否存在
	if row.FulfillmentRuleID <= 0 {
		return true, "提货卡模式商品未绑定发卡规则"
	}

	// 查询发卡规则
	var rule fulfillmentRuleRow
	err := db.WithContext(ctx).
		Table("sms_product_fulfillment_rule").
		Select("id, rule_status, card_template_id, platform_id, tenant_id, merchant_id").
		Where("id = ? AND is_deleted = 0", row.FulfillmentRuleID).
		Take(&rule).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return true, fmt.Sprintf("发卡规则[%d]不存在或已删除", row.FulfillmentRuleID)
		}
		return true, fmt.Sprintf("查询发卡规则失败: %v", err)
	}

	// 检查发卡规则状态
	if rule.RuleStatus != 1 {
		return true, fmt.Sprintf("发卡规则[%d]已禁用", row.FulfillmentRuleID)
	}

	// 检查作用域一致性
	if rule.PlatformID != row.PlatformID || rule.TenantID != row.TenantID || rule.MerchantID != row.MerchantID {
		return true, "发卡规则的作用域与商品不一致"
	}

	// 检查卡片模板
	if rule.CardTemplateID > 0 {
		var template cardTemplateRow
		err := db.WithContext(ctx).
			Table("sms_card_template").
			Select("id, status").
			Where("id = ? AND is_deleted = 0", rule.CardTemplateID).
			Take(&template).Error
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return true, fmt.Sprintf("卡片模板[%d]不存在或已删除", rule.CardTemplateID)
			}
			return true, fmt.Sprintf("查询卡片模板失败: %v", err)
		}
		if template.Status == 0 {
			return true, fmt.Sprintf("卡片模板[%d]已禁用", rule.CardTemplateID)
		}
	}

	return false, ""
}

// ProductConflictStatus 商品冲突状态
type ProductConflictStatus struct {
	HasConflict bool   `json:"hasConflict"`
	ConflictMsg string `json:"conflictMsg"`
}

// BatchCheckProductFulfillmentConflict 批量检查商品履约模式配置冲突
func BatchCheckProductFulfillmentConflict(ctx context.Context, db *gorm.DB, ids []int64) (map[int64]*ProductConflictStatus, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	rows, err := LoadProductVisibilityRows(ctx, db, pkgscope.DefaultScope(0, 0, 0), ids)
	if err != nil {
		return nil, err
	}

	result := make(map[int64]*ProductConflictStatus, len(rows))
	for _, row := range rows {
		hasConflict, conflictMsg := CheckProductFulfillmentConflict(ctx, db, row)
		result[row.ID] = &ProductConflictStatus{
			HasConflict: hasConflict,
			ConflictMsg: conflictMsg,
		}
	}

	return result, nil
}
