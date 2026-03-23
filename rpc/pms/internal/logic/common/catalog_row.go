package common

import (
	"time"

	pkgscope "github.com/feihua/zero-admin/pkg/scope"
)

type CatalogScopeRow struct {
	ID                  int64      `gorm:"column:id"`
	ParentID            int64      `gorm:"column:parent_id"`
	Name                string     `gorm:"column:name"`
	Logo                string     `gorm:"column:logo"`
	BigPic              string     `gorm:"column:big_pic"`
	Description         string     `gorm:"column:description"`
	FirstLetter         string     `gorm:"column:first_letter"`
	Sort                int32      `gorm:"column:sort"`
	RecommendStatus     int32      `gorm:"column:recommend_status"`
	ProductCount        int32      `gorm:"column:product_count"`
	ProductCommentCount int32      `gorm:"column:product_comment_count"`
	ProductUnit         string     `gorm:"column:product_unit"`
	NavStatus           int32      `gorm:"column:nav_status"`
	Icon                string     `gorm:"column:icon"`
	Keywords            string     `gorm:"column:keywords"`
	Level               int32      `gorm:"column:level"`
	IsEnabled           int32      `gorm:"column:is_enabled"`
	CreateBy            int64      `gorm:"column:create_by"`
	CreateTime          time.Time  `gorm:"column:create_time"`
	UpdateBy            *int64     `gorm:"column:update_by"`
	UpdateTime          *time.Time `gorm:"column:update_time"`
	PlatformID          int64      `gorm:"column:platform_id"`
	TenantID            int64      `gorm:"column:tenant_id"`
	MerchantID          int64      `gorm:"column:merchant_id"`
}

func (r CatalogScopeRow) GovernanceScope() pkgscope.GovernanceScope {
	scope, _ := pkgscope.NormalizeGovernanceScope("", r.PlatformID, r.TenantID, r.MerchantID)
	return scope
}
