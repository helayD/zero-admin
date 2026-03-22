package merchantmodel

import "time"

const (
	MerchantReviewPending            int32 = 0
	MerchantReviewMaterialRequired   int32 = 1
	MerchantReviewRejected           int32 = 2
	MerchantReviewApproved           int32 = 3
	MerchantBusinessPendingActivation int32 = 0
	MerchantBusinessEnabled          int32 = 1
	MerchantBusinessDisabled         int32 = 2
	MerchantBusinessArchived         int32 = 3
)

type SysMerchant struct {
	ID                 int64      `gorm:"column:id;primaryKey;autoIncrement:true"`
	PlatformID         int64      `gorm:"column:platform_id"`
	TenantID           int64      `gorm:"column:tenant_id"`
	MerchantCode       string     `gorm:"column:merchant_code"`
	MerchantName       string     `gorm:"column:merchant_name"`
	MerchantShortName  string     `gorm:"column:merchant_short_name"`
	ContactName        string     `gorm:"column:contact_name"`
	ContactMobile      string     `gorm:"column:contact_mobile"`
	ContactEmail       string     `gorm:"column:contact_email"`
	AvailableChannels  string     `gorm:"column:available_channels"`
	CapabilityFlags    string     `gorm:"column:capability_flags"`
	ReviewStatus       int32      `gorm:"column:review_status"`
	ReviewReason       string     `gorm:"column:review_reason"`
	ReviewedBy         int64      `gorm:"column:reviewed_by"`
	ReviewedByName     string     `gorm:"column:reviewed_by_name"`
	ReviewedAt         *time.Time `gorm:"column:reviewed_at"`
	BusinessStatus     int32      `gorm:"column:business_status"`
	StatusReason       string     `gorm:"column:status_reason"`
	VisibleScopeHint   string     `gorm:"column:visible_scope_hint"`
	PrimaryAdminUserID int64      `gorm:"column:primary_admin_user_id"`
	Remark             string     `gorm:"column:remark"`
	ArchivedAt         *time.Time `gorm:"column:archived_at"`
	CreatedBy          string     `gorm:"column:created_by"`
	CreatedAt          time.Time  `gorm:"column:created_at"`
	UpdatedBy          string     `gorm:"column:updated_by"`
	UpdatedAt          *time.Time `gorm:"column:updated_at"`
}

func (*SysMerchant) TableName() string {
	return "sys_merchant"
}

type SysMerchantAudit struct {
	ID                   int64      `gorm:"column:id;primaryKey;autoIncrement:true"`
	TraceID              string     `gorm:"column:trace_id"`
	PlatformID           int64      `gorm:"column:platform_id"`
	TenantID             int64      `gorm:"column:tenant_id"`
	MerchantID           int64      `gorm:"column:merchant_id"`
	MerchantCode         string     `gorm:"column:merchant_code"`
	Action               string     `gorm:"column:action"`
	BeforeReviewStatus   *int32     `gorm:"column:before_review_status"`
	AfterReviewStatus    *int32     `gorm:"column:after_review_status"`
	BeforeBusinessStatus *int32     `gorm:"column:before_business_status"`
	AfterBusinessStatus  *int32     `gorm:"column:after_business_status"`
	OperatorID           int64      `gorm:"column:operator_id"`
	OperatorName         string     `gorm:"column:operator_name"`
	RequestPayload       string     `gorm:"column:request_payload"`
	EventPayload         string     `gorm:"column:event_payload"`
	Result               string     `gorm:"column:result"`
	CreatedBy            string     `gorm:"column:created_by"`
	CreatedAt            time.Time  `gorm:"column:created_at"`
	UpdatedBy            string     `gorm:"column:updated_by"`
	UpdatedAt            *time.Time `gorm:"column:updated_at"`
}

func (*SysMerchantAudit) TableName() string {
	return "sys_merchant_audit"
}
