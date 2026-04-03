package model

import "time"

// ProductCommentAuditLog 商品评价审核日志。
type ProductCommentAuditLog struct {
	ID           int64     `gorm:"column:id;primaryKey;autoIncrement:true" json:"id"`
	CommentID    string    `gorm:"column:comment_id;size:32;not null;index" json:"comment_id"`
	PlatformID   int64     `gorm:"column:platform_id;not null;default:0" json:"platform_id"`
	TenantID     int64     `gorm:"column:tenant_id;not null;default:0;index" json:"tenant_id"`
	MerchantID   int64     `gorm:"column:merchant_id;not null;default:0;index" json:"merchant_id"`
	Action       string    `gorm:"column:action;size:32;not null;default:''" json:"action"`
	FromStatus   int32     `gorm:"column:from_status;not null;default:0" json:"from_status"`
	ToStatus     int32     `gorm:"column:to_status;not null;default:0" json:"to_status"`
	OperatorID   int64     `gorm:"column:operator_id;not null;default:0" json:"operator_id"`
	OperatorName string    `gorm:"column:operator_name;size:64;default:''" json:"operator_name"`
	Remark       string    `gorm:"column:remark;size:500;default:''" json:"remark"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime" json:"created_at"`
}

func (ProductCommentAuditLog) TableName() string {
	return "pms_comment_audit_log"
}
