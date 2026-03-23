package productcategoryservicelogic

import (
	pkgscope "github.com/feihua/zero-admin/pkg/scope"
	"gorm.io/gorm"
)

func replaceCategoryRelations(tx *gorm.DB, categoryID int64, attrIDs []int64, current pkgscope.GovernanceScope) error {
	if err := tx.Table("pms_product_category_attribute_relation").Where("product_category_id = ?", categoryID).Delete(nil).Error; err != nil {
		return err
	}
	for _, attrID := range attrIDs {
		if attrID <= 0 {
			continue
		}
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
