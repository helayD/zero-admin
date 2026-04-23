package digitalcardmint

import "time"

type CardMintTaskRow struct {
	ID                    int64      `gorm:"column:id"`
	PlatformID            int64      `gorm:"column:platform_id"`
	TenantID              int64      `gorm:"column:tenant_id"`
	MerchantID            int64      `gorm:"column:merchant_id"`
	AssetInstanceID       int64      `gorm:"column:asset_instance_id"`
	ParticipationRecordID int64      `gorm:"column:participation_record_id"`
	ActivityID            int64      `gorm:"column:activity_id"`
	MemberID              int64      `gorm:"column:member_id"`
	RequestID             string     `gorm:"column:request_id"`
	TraceID               string     `gorm:"column:trace_id"`
	IdempotencyKey        string     `gorm:"column:idempotency_key"`
	TaskStatus            string     `gorm:"column:task_status"`
	MintStatus            string     `gorm:"column:mint_status"`
	ChainStatus           string     `gorm:"column:chain_status"`
	TokenID               string     `gorm:"column:token_id"`
	ChainTxID             string     `gorm:"column:chain_tx_id"`
	RetryCount            int32      `gorm:"column:retry_count"`
	MaxRetryCount         int32      `gorm:"column:max_retry_count"`
	LastErrorCode         string     `gorm:"column:last_error_code"`
	LastErrorReason       string     `gorm:"column:last_error_reason"`
	LastReceiptSummary    string     `gorm:"column:last_receipt_summary"`
	LastReceiptJSON       string     `gorm:"column:last_receipt_json"`
	LastExecuteAt         *time.Time `gorm:"column:last_execute_at"`
	NextRetryAt           *time.Time `gorm:"column:next_retry_at"`
	ManualRequired        int32      `gorm:"column:manual_required"`
	Frozen                int32      `gorm:"column:frozen"`
	FreezeReason          string     `gorm:"column:freeze_reason"`
	CreateBy              int64      `gorm:"column:create_by"`
	UpdateBy              int64      `gorm:"column:update_by"`
	CreateTime            time.Time  `gorm:"column:create_time"`
	UpdateTime            *time.Time `gorm:"column:update_time"`
	IsDeleted             int32      `gorm:"column:is_deleted"`
}

func (CardMintTaskRow) TableName() string {
	return "sms_card_mint_task"
}

type CardInstanceRow struct {
	ID                    int64      `gorm:"column:id"`
	PlatformID            int64      `gorm:"column:platform_id"`
	TenantID              int64      `gorm:"column:tenant_id"`
	MerchantID            int64      `gorm:"column:merchant_id"`
	ActivityID            int64      `gorm:"column:activity_id"`
	MemberID              int64      `gorm:"column:member_id"`
	ParticipationRecordID int64      `gorm:"column:participation_record_id"`
	RequestID             string     `gorm:"column:request_id"`
	TraceID               string     `gorm:"column:trace_id"`
	Scope                 string     `gorm:"column:scope"`
	PoolID                int64      `gorm:"column:pool_id"`
	TemplateID            int64      `gorm:"column:template_id"`
	Rarity                string     `gorm:"column:rarity"`
	AssetNo               string     `gorm:"column:asset_no"`
	AssetStatus           string     `gorm:"column:asset_status"`
	MintStatus            string     `gorm:"column:mint_status"`
	TokenID               string     `gorm:"column:token_id"`
	ChainStatus           string     `gorm:"column:chain_status"`
	LastReceiptAt         *time.Time `gorm:"column:last_receipt_at"`
	MintTaskID            int64      `gorm:"column:mint_task_id"`
	IssuedAt              *time.Time `gorm:"column:issued_at"`
	CreateBy              int64      `gorm:"column:create_by"`
	CreateTime            *time.Time `gorm:"column:create_time"`
	UpdateBy              *int64     `gorm:"column:update_by"`
	UpdateTime            *time.Time `gorm:"column:update_time"`
	IsDeleted             int32      `gorm:"column:is_deleted"`
}

func (CardInstanceRow) TableName() string {
	return "sms_card_instance"
}

type CardAssetLogRow struct {
	ID                    int64     `gorm:"column:id"`
	AssetInstanceID       int64     `gorm:"column:asset_instance_id"`
	ParticipationRecordID int64     `gorm:"column:participation_record_id"`
	FromStatus            string    `gorm:"column:from_status"`
	ToStatus              string    `gorm:"column:to_status"`
	OperationType         string    `gorm:"column:operation_type"`
	OperatorType          string    `gorm:"column:operator_type"`
	TraceID               string    `gorm:"column:trace_id"`
	ReasonCode            string    `gorm:"column:reason_code"`
	ReasonText            string    `gorm:"column:reason_text"`
	PayloadJSON           string    `gorm:"column:payload_json"`
	CreateTime            time.Time `gorm:"column:create_time"`
}

func (CardAssetLogRow) TableName() string {
	return "sms_card_asset_log"
}

type ParticipationRecordRow struct {
	ID                  int64  `gorm:"column:id"`
	ActivityID          int64  `gorm:"column:activity_id"`
	MemberID            int64  `gorm:"column:member_id"`
	PoolID              int64  `gorm:"column:pool_id"`
	TemplateID          int64  `gorm:"column:template_id"`
	RequestID           string `gorm:"column:request_id"`
	TraceID             string `gorm:"column:trace_id"`
	EligibilitySnapshot string `gorm:"column:eligibility_snapshot_json"`
	AssetInstanceID     int64  `gorm:"column:asset_instance_id"`
	IsDeleted           int32  `gorm:"column:is_deleted"`
}

func (ParticipationRecordRow) TableName() string {
	return "sms_draw_participation_record"
}

type DrawActivityRow struct {
	ID                 int64  `gorm:"column:id"`
	Name               string `gorm:"column:name"`
	PlatformID         int64  `gorm:"column:platform_id"`
	TenantID           int64  `gorm:"column:tenant_id"`
	MerchantID         int64  `gorm:"column:merchant_id"`
	RealNameRequired   int32  `gorm:"column:real_name_required"`
	PublishReadiness   int32  `gorm:"column:publish_readiness"`
	CopyrightStatus    int32  `gorm:"column:copyright_status"`
	ContentAuditStatus int32  `gorm:"column:content_audit_status"`
	Status             int32  `gorm:"column:status"`
	AuditStatus        int32  `gorm:"column:audit_status"`
	IsEnabled          int32  `gorm:"column:is_enabled"`
	IsDeleted          int32  `gorm:"column:is_deleted"`
}

func (DrawActivityRow) TableName() string {
	return "sms_draw_activity"
}

type CardTemplateRow struct {
	ID                 int64  `gorm:"column:id"`
	TemplateName       string `gorm:"column:template_name"`
	DisplayStatus      int32  `gorm:"column:display_status"`
	ContentAuditStatus int32  `gorm:"column:content_audit_status"`
	CredentialRef      string `gorm:"column:credential_ref"`
	Status             int32  `gorm:"column:status"`
	AuditStatus        int32  `gorm:"column:audit_status"`
	IsDeleted          int32  `gorm:"column:is_deleted"`
}

func (CardTemplateRow) TableName() string {
	return "sms_card_template"
}

type MintRequestedEvent struct {
	EventName       string `json:"eventName"`
	TaskID          int64  `json:"taskId"`
	AssetInstanceID int64  `json:"assetInstanceId"`
	RequestID       string `json:"requestId"`
	TraceID         string `json:"traceId"`
}

type QueryFilter struct {
	PageNum        int32
	PageSize       int32
	ActivityID     int64
	ActivityName   string
	MemberID       int64
	TemplateID     int64
	AssetNo        string
	TokenID        string
	TaskStatus     string
	MintStatus     string
	ChainStatus    string
	ManualRequired int32
	StartTime      string
	EndTime        string
}

type TaskListItem struct {
	TaskID                int64  `json:"taskId"`
	TraceID               string `json:"traceId"`
	AssetInstanceID       int64  `json:"assetInstanceId"`
	AssetNo               string `json:"assetNo"`
	MemberID              int64  `json:"memberId"`
	ActivityID            int64  `json:"activityId"`
	ActivityName          string `json:"activityName"`
	TemplateID            int64  `json:"templateId"`
	TemplateName          string `json:"templateName"`
	TokenID               string `json:"tokenId"`
	TaskStatus            string `json:"taskStatus"`
	TaskStatusText        string `json:"taskStatusText"`
	MintStatus            string `json:"mintStatus"`
	MintStatusText        string `json:"mintStatusText"`
	ChainStatus           string `json:"chainStatus"`
	ChainStatusText       string `json:"chainStatusText"`
	RetryCount            int32  `json:"retryCount"`
	LastError             string `json:"lastError"`
	LastExecuteAt         string `json:"lastExecuteAt"`
	ManualRequired        bool   `json:"manualRequired"`
	Frozen                bool   `json:"frozen"`
	FreezeReason          string `json:"freezeReason"`
	LastReceiptSummary    string `json:"lastReceiptSummary"`
	AssetStatusText       string `json:"assetStatusText"`
	ParticipationRecordID int64  `json:"participationRecordId"`
	ChainType             string `json:"chainType"`
}

type AssetLogItem struct {
	OperationType string `json:"operationType"`
	OperatorType  string `json:"operatorType"`
	FromStatus    string `json:"fromStatus"`
	ToStatus      string `json:"toStatus"`
	ReasonText    string `json:"reasonText"`
	TraceID       string `json:"traceId"`
	PayloadJSON   string `json:"payloadJson"`
	CreateTime    string `json:"createTime"`
}

type TaskDetail struct {
	Item            TaskListItem   `json:"item"`
	RequestID       string         `json:"requestId"`
	ChainTxID       string         `json:"chainTxId"`
	LastReceiptJSON string         `json:"lastReceiptJson"`
	Logs            []AssetLogItem `json:"logs"`
	ChainType       string         `json:"chainType"`
}

type ActionResult struct {
	TaskID         int64
	TraceID        string
	TaskStatus     string
	MintStatus     string
	ChainStatus    string
	RetryCount     int32
	ManualRequired bool
	Frozen         bool
	ChainType      string
}

type ExecuteResult struct {
	TaskID         int64
	TokenID        string
	TaskStatus     string
	MintStatus     string
	ChainStatus    string
	RetryCount     int32
	ManualRequired bool
}

type RecoveryStats struct {
	Dispatched int32
	Executed   int32
	Escalated  int32
}
