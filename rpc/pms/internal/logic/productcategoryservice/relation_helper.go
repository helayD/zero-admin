package productcategoryservicelogic

import (
	"strings"

	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	logiccommon "github.com/feihua/zero-admin/rpc/pms/internal/logic/common"
	"gorm.io/gorm"
)

func replaceCategoryRelations(tx *gorm.DB, categoryID int64, attrIDs []int64, current pkgscope.GovernanceScope) error {
	validatedIDs, err := logiccommon.ValidateCategoryAttributeBindings(tx.Statement.Context, tx, current, attrIDs)
	if err != nil {
		return err
	}
	if err := tx.Table("pms_product_category_attribute_relation").Where("product_category_id = ?", categoryID).Delete(nil).Error; err != nil {
		return err
	}
	for _, attrID := range validatedIDs {
		if err := tx.Table("pms_product_category_attribute_relation").Create(map[string]interface{}{
			"product_category_id":  categoryID,
			"product_attribute_id": attrID,
			"platform_id":          current.PlatformID,
			"tenant_id":            current.TenantID,
			"merchant_id":          current.MerchantID,
		}).Error; err != nil {
			return err
		}
	}
	return nil
}

func isCategoryBindingValidationError(err error) bool {
	if err == nil {
		return false
	}
	message := err.Error()
	return strings.Contains(message, "绑定的商品属性") || strings.Contains(message, "只能绑定启用状态")
}
