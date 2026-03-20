package tenantmodel

import "time"

const (
	TenantStatusDisabled          int32 = 0
	TenantStatusEnabled           int32 = 1
	TenantStatusPendingActivation int32 = 2
	TenantStatusArchived          int32 = 3
)

type SysTenant struct {
	ID                 int64      `gorm:"column:id;primaryKey;autoIncrement:true"`
	PlatformID         int64      `gorm:"column:platform_id"`
	TenantCode         string     `gorm:"column:tenant_code"`
	TenantName         string     `gorm:"column:tenant_name"`
	TenantShortName    string     `gorm:"column:tenant_short_name"`
	ContactName        string     `gorm:"column:contact_name"`
	ContactMobile      string     `gorm:"column:contact_mobile"`
	ContactEmail       string     `gorm:"column:contact_email"`
	AvailableChannels  string     `gorm:"column:available_channels"`
	DataRetentionDays  int32      `gorm:"column:data_retention_days"`
	FeatureFlags       string     `gorm:"column:feature_flags"`
	Status             int32      `gorm:"column:status"`
	StatusReason       string     `gorm:"column:status_reason"`
	ArchivedAt         *time.Time `gorm:"column:archived_at"`
	CreatedBy          string     `gorm:"column:created_by"`
	CreatedAt          time.Time  `gorm:"column:created_at"`
	UpdatedBy          string     `gorm:"column:updated_by"`
	UpdatedAt          *time.Time `gorm:"column:updated_at"`
}

func (*SysTenant) TableName() string {
	return "sys_tenant"
}

type SysTenantUser struct {
	ID               int64      `gorm:"column:id;primaryKey;autoIncrement:true"`
	PlatformID       int64      `gorm:"column:platform_id"`
	TenantID         int64      `gorm:"column:tenant_id"`
	UserID           int64      `gorm:"column:user_id"`
	RoleMode         string     `gorm:"column:role_mode"`
	ActivationStatus int32      `gorm:"column:activation_status"`
	IsPrimaryAdmin   int32      `gorm:"column:is_primary_admin"`
	ScopeMetadata    string     `gorm:"column:scope_metadata"`
	CreatedBy        string     `gorm:"column:created_by"`
	CreatedAt        time.Time  `gorm:"column:created_at"`
	UpdatedBy        string     `gorm:"column:updated_by"`
	UpdatedAt        *time.Time `gorm:"column:updated_at"`
}

func (*SysTenantUser) TableName() string {
	return "sys_tenant_user"
}

type SysTenantAudit struct {
	ID             int64      `gorm:"column:id;primaryKey;autoIncrement:true"`
	PlatformID     int64      `gorm:"column:platform_id"`
	TenantID       int64      `gorm:"column:tenant_id"`
	TenantCode     string     `gorm:"column:tenant_code"`
	Action         string     `gorm:"column:action"`
	BeforeStatus   *int32     `gorm:"column:before_status"`
	AfterStatus    int32      `gorm:"column:after_status"`
	OperatorID     int64      `gorm:"column:operator_id"`
	OperatorName   string     `gorm:"column:operator_name"`
	RequestPayload string     `gorm:"column:request_payload"`
	EventPayload   string     `gorm:"column:event_payload"`
	Result         string     `gorm:"column:result"`
	CreatedBy      string     `gorm:"column:created_by"`
	CreatedAt      time.Time  `gorm:"column:created_at"`
	UpdatedBy      string     `gorm:"column:updated_by"`
	UpdatedAt      *time.Time `gorm:"column:updated_at"`
}

func (*SysTenantAudit) TableName() string {
	return "sys_tenant_audit"
}
