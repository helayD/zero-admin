package types

type QueryDigitalCardChainListReq struct {
	ScopeType      string `form:"scopeType,optional" json:"scopeType"`
	PlatformId     int64  `form:"platformId,optional" json:"platformId"`
	TenantId       int64  `form:"tenantId,optional" json:"tenantId"`
	MerchantId     int64  `form:"merchantId,optional" json:"merchantId"`
	PageSize       int    `form:"pageSize,default=20" json:"pageSize"`
	Current        int    `form:"current,default=1" json:"current"`
	ActivityId     int64  `form:"activityId,optional" json:"activityId"`
	ActivityName   string `form:"activityName,optional" json:"activityName"`
	MemberId       int64  `form:"memberId,optional" json:"memberId"`
	TemplateId     int64  `form:"templateId,optional" json:"templateId"`
	AssetNo        string `form:"assetNo,optional" json:"assetNo"`
	TokenId        string `form:"tokenId,optional" json:"tokenId"`
	TaskStatus     string `form:"taskStatus,optional" json:"taskStatus"`
	MintStatus     string `form:"mintStatus,optional" json:"mintStatus"`
	ChainStatus    string `form:"chainStatus,optional" json:"chainStatus"`
	ManualRequired int32  `form:"manualRequired,optional" json:"manualRequired"`
	StartTime      string `form:"startTime,optional" json:"startTime"`
	EndTime        string `form:"endTime,optional" json:"endTime"`
}

type DigitalCardChainItem struct {
	TaskId                int64  `json:"taskId"`
	TraceId               string `json:"traceId"`
	AssetInstanceId       int64  `json:"assetInstanceId"`
	AssetNo               string `json:"assetNo"`
	MemberId              int64  `json:"memberId"`
	ActivityId            int64  `json:"activityId"`
	ActivityName          string `json:"activityName"`
	TemplateId            int64  `json:"templateId"`
	TemplateName          string `json:"templateName"`
	TokenId               string `json:"tokenId"`
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
	ParticipationRecordId int64  `json:"participationRecordId"`
}

type QueryDigitalCardChainListData struct {
	List  []*DigitalCardChainItem `json:"list"`
	Total int64                   `json:"total"`
}

type QueryDigitalCardChainListResp struct {
	Code     string                        `json:"code"`
	Message  string                        `json:"message"`
	Data     QueryDigitalCardChainListData `json:"data"`
	Current  int                           `json:"current"`
	PageSize int                           `json:"pageSize"`
	Total    int64                         `json:"total"`
	Success  bool                          `json:"success"`
}

type QueryDigitalCardChainDetailReq struct {
	TaskId int64 `form:"taskId" json:"taskId"`
}

type DigitalCardChainLogItem struct {
	OperationType string `json:"operationType"`
	OperatorType  string `json:"operatorType"`
	FromStatus    string `json:"fromStatus"`
	ToStatus      string `json:"toStatus"`
	ReasonText    string `json:"reasonText"`
	TraceId       string `json:"traceId"`
	PayloadJson   string `json:"payloadJson"`
	CreateTime    string `json:"createTime"`
}

type DigitalCardChainDetailData struct {
	Item            DigitalCardChainItem      `json:"item"`
	RequestId       string                    `json:"requestId"`
	ChainTxId       string                    `json:"chainTxId"`
	LastReceiptJson string                    `json:"lastReceiptJson"`
	Logs            []DigitalCardChainLogItem `json:"logs"`
}

type QueryDigitalCardChainDetailResp struct {
	Code    string                     `json:"code"`
	Message string                     `json:"message"`
	Data    DigitalCardChainDetailData `json:"data"`
	Success bool                       `json:"success"`
}

type RetryDigitalCardChainReq struct {
	TaskId     int64  `json:"taskId"`
	Reason     string `json:"reason" validate:"required"`
	ScopeType  string `json:"scopeType,optional"`
	PlatformId int64  `json:"platformId,optional"`
	TenantId   int64  `json:"tenantId,optional"`
	MerchantId int64  `json:"merchantId,optional"`
}

type FreezeDigitalCardChainReq = RetryDigitalCardChainReq
type EscalateDigitalCardChainReq = RetryDigitalCardChainReq

type DigitalCardChainActionResp struct {
	Code           string `json:"code"`
	Message        string `json:"message"`
	TaskStatus     string `json:"taskStatus"`
	MintStatus     string `json:"mintStatus"`
	ChainStatus    string `json:"chainStatus"`
	RetryCount     int32  `json:"retryCount"`
	ManualRequired bool   `json:"manualRequired"`
	Frozen         bool   `json:"frozen"`
	Success        bool   `json:"success"`
}

type QueryDigitalCardChainActionsReq struct {
	TaskId int64 `form:"taskId" json:"taskId"`
}

type QueryDigitalCardChainActionsResp struct {
	Code             string   `json:"code"`
	Message          string   `json:"message"`
	AvailableActions []string `json:"availableActions"`
	Success          bool     `json:"success"`
}
