package productfulfillmentruleservice

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

// validateCardTemplateForRule 校验发卡规则关联的卡片模板的合法性。
//
// Story 10.10 Task 8.1 + 修复 M5: 在新增/更新发卡规则前，必须确保
//  1. 卡片模板存在 (sms_card_template, is_deleted=0)
//  2. 模板未被禁用 (status=1)
//  3. 模板的 scope 对当前请求 scope 可见
//
// Scope 下放规则（与 Story AC1 描述一致）：
//   - 平台模板 (tenant_id=0, merchant_id=0)：同 platform_id 下所有 scope 可引用
//   - 租户模板 (tenant_id>0, merchant_id=0)：同 platform_id+tenant_id 下任意 merchant 可引用
//   - 商户模板 (tenant_id>0, merchant_id>0)：仅同 platform_id+tenant_id+merchant_id 可引用
func validateCardTemplateForRule(ctx context.Context, db *gorm.DB, cardTemplateId, platformId, tenantId, merchantId int64) error {
	if cardTemplateId <= 0 {
		return errors.New("关联卡片模板ID无效")
	}

	type tplRow struct {
		Id           int64  `gorm:"column:id"`
		TemplateName string `gorm:"column:template_name"`
		Status       int32  `gorm:"column:status"`
		PlatformId   int64  `gorm:"column:platform_id"`
		TenantId     int64  `gorm:"column:tenant_id"`
		MerchantId   int64  `gorm:"column:merchant_id"`
	}

	var row tplRow
	err := db.WithContext(ctx).
		Table("sms_card_template").
		Select("id, template_name, status, platform_id, tenant_id, merchant_id").
		Where("id = ? AND is_deleted = 0", cardTemplateId).
		Take(&row).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("卡片模板[ID:%d]不存在或已被删除", cardTemplateId)
		}
		return fmt.Errorf("校验卡片模板失败: %w", err)
	}

	if row.Status == 0 {
		return fmt.Errorf("卡片模板[%s]已被禁用，请先启用或更换", row.TemplateName)
	}

	// platformId 必须一致
	if row.PlatformId != platformId {
		return fmt.Errorf("卡片模板[%s]不属于当前平台，无权引用", row.TemplateName)
	}

	// 平台模板：tenant=0 且 merchant=0 → 同平台任何 scope 可引用
	if row.TenantId == 0 && row.MerchantId == 0 {
		return nil
	}
	// 租户模板：tenant>0 且 merchant=0 → 同租户下任意 merchant 可引用
	if row.MerchantId == 0 && row.TenantId == tenantId {
		return nil
	}
	// 商户模板：精确匹配
	if row.TenantId == tenantId && row.MerchantId == merchantId {
		return nil
	}

	return fmt.Errorf("卡片模板[%s]不属于当前作用域，无权引用", row.TemplateName)
}
