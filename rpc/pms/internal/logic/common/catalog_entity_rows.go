package common

import "time"

type ProductAttributeRow struct {
	ID           int64      `gorm:"column:id"`
	GroupID      int64      `gorm:"column:group_id"`
	Name         string     `gorm:"column:name"`
	InputType    int32      `gorm:"column:input_type"`
	ValueType    int32      `gorm:"column:value_type"`
	InputList    string     `gorm:"column:input_list"`
	Unit         string     `gorm:"column:unit"`
	IsRequired   int32      `gorm:"column:is_required"`
	IsSearchable int32      `gorm:"column:is_searchable"`
	IsShow       int32      `gorm:"column:is_show"`
	Sort         int32      `gorm:"column:sort"`
	Status       int32      `gorm:"column:status"`
	CreateBy     int64      `gorm:"column:create_by"`
	CreateTime   time.Time  `gorm:"column:create_time"`
	UpdateBy     *int64     `gorm:"column:update_by"`
	UpdateTime   *time.Time `gorm:"column:update_time"`
	IsDeleted    int32      `gorm:"column:is_deleted"`
	PlatformID   int64      `gorm:"column:platform_id"`
	TenantID     int64      `gorm:"column:tenant_id"`
	MerchantID   int64      `gorm:"column:merchant_id"`
}

type ProductAttributeGroupRow struct {
	ID         int64      `gorm:"column:id"`
	CategoryID int64      `gorm:"column:category_id"`
	Name       string     `gorm:"column:name"`
	Sort       int32      `gorm:"column:sort"`
	Status     int32      `gorm:"column:status"`
	CreateBy   int64      `gorm:"column:create_by"`
	CreateTime time.Time  `gorm:"column:create_time"`
	UpdateBy   *int64     `gorm:"column:update_by"`
	UpdateTime *time.Time `gorm:"column:update_time"`
	IsDeleted  int32      `gorm:"column:is_deleted"`
	PlatformID int64      `gorm:"column:platform_id"`
	TenantID   int64      `gorm:"column:tenant_id"`
	MerchantID int64      `gorm:"column:merchant_id"`
}

type ProductSpecRow struct {
	ID         int64      `gorm:"column:id"`
	CategoryID int64      `gorm:"column:category_id"`
	Name       string     `gorm:"column:name"`
	Sort       int32      `gorm:"column:sort"`
	Status     int32      `gorm:"column:status"`
	CreateBy   int64      `gorm:"column:create_by"`
	CreateTime time.Time  `gorm:"column:create_time"`
	UpdateBy   *int64     `gorm:"column:update_by"`
	UpdateTime *time.Time `gorm:"column:update_time"`
	IsDeleted  int32      `gorm:"column:is_deleted"`
	PlatformID int64      `gorm:"column:platform_id"`
	TenantID   int64      `gorm:"column:tenant_id"`
	MerchantID int64      `gorm:"column:merchant_id"`
}

type ProductSpecValueRow struct {
	ID         int64      `gorm:"column:id"`
	SpecID      int64      `gorm:"column:spec_id"`
	Value      string     `gorm:"column:value"`
	Sort       int32      `gorm:"column:sort"`
	Status     int32      `gorm:"column:status"`
	CreateBy   int64      `gorm:"column:create_by"`
	CreateTime time.Time  `gorm:"column:create_time"`
	UpdateBy   *int64     `gorm:"column:update_by"`
	UpdateTime *time.Time `gorm:"column:update_time"`
	IsDeleted  int32      `gorm:"column:is_deleted"`
	PlatformID int64      `gorm:"column:platform_id"`
	TenantID   int64      `gorm:"column:tenant_id"`
	MerchantID int64      `gorm:"column:merchant_id"`
}
