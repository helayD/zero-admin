package types

type DigitalCardAssetGovernanceScopeReq struct {
	ScopeType  string `form:"scopeType,optional" json:"scopeType,optional"`
	PlatformId int64  `form:"platformId,optional" json:"platformId,optional"`
	TenantId   int64  `form:"tenantId,optional" json:"tenantId,optional"`
	MerchantId int64  `form:"merchantId,optional" json:"merchantId,optional"`
}

type DigitalCardAssetQueryGovernanceScopeReq struct {
	ScopeType  string `form:"scopeType,optional"`
	PlatformId int64  `form:"platformId,optional"`
	TenantId   int64  `form:"tenantId,optional"`
	MerchantId int64  `form:"merchantId,optional"`
}

type QueryDigitalCardAssetListReq struct {
	DigitalCardAssetQueryGovernanceScopeReq
	PageSize         int    `form:"pageSize,default=20"`
	Current          int    `form:"current,default=1"`
	ActivityId       int64  `form:"activityId,optional"`
	ActivityName     string `form:"activityName,optional"`
	MemberId         int64  `form:"memberId,optional"`
	TemplateId       int64  `form:"templateId,optional"`
	TemplateName     string `form:"templateName,optional"`
	AssetNo          string `form:"assetNo,optional"`
	TokenId          string `form:"tokenId,optional"`
	MintStatus       string `form:"mintStatus,optional"`
	ChainStatus      string `form:"chainStatus,optional"`
	DisplayStatus    string `form:"displayStatus,optional"`
	ComplianceStatus string `form:"complianceStatus,optional"`
	StartTime        string `form:"startTime,optional"`
	EndTime          string `form:"endTime,optional"`
}

type DigitalCardAssetItem struct {
	AssetInstanceId       int64  `json:"assetInstanceId"`
	AssetNo               string `json:"assetNo"`
	MemberId              int64  `json:"memberId"`
	ActivityId            int64  `json:"activityId"`
	ActivityName          string `json:"activityName"`
	TemplateId            int64  `json:"templateId"`
	TemplateName          string `json:"templateName"`
	CardFaceImage         string `json:"cardFaceImage"`
	Rarity                string `json:"rarity"`
	TokenId               string `json:"tokenId"`
	MintTaskId            int64  `json:"mintTaskId"`
	MintStatus            string `json:"mintStatus"`
	MintStatusText        string `json:"mintStatusText"`
	ChainStatus           string `json:"chainStatus"`
	ChainStatusText       string `json:"chainStatusText"`
	DisplayStatus         string `json:"displayStatus"`
	DisplayStatusText     string `json:"displayStatusText"`
	ComplianceStatus      string `json:"complianceStatus"`
	ComplianceStatusText  string `json:"complianceStatusText"`
	TokenStatusText       string `json:"tokenStatusText"`
	ComplianceRuleSummary string `json:"complianceRuleSummary"`
	LastReceiptSummary    string `json:"lastReceiptSummary"`
	ObtainedAt            string `json:"obtainedAt"`
	DisposedAt            string `json:"disposedAt"`
	LatestReasonSummary   string `json:"latestReasonSummary"`
	ChainType             string `json:"chainType"`
}

type QueryDigitalCardAssetListData struct {
	List  []*DigitalCardAssetItem `json:"list"`
	Total int64                   `json:"total"`
}

type QueryDigitalCardAssetListResp struct {
	Code     string                        `json:"code"`
	Message  string                        `json:"message"`
	Data     QueryDigitalCardAssetListData `json:"data"`
	Current  int                           `json:"current"`
	PageSize int                           `json:"pageSize"`
	Total    int64                         `json:"total"`
	Success  bool                          `json:"success"`
}

type QueryDigitalCardAssetDetailReq struct {
	AssetInstanceId int64 `form:"assetInstanceId"`
}

type DigitalCardAssetParticipationSummary struct {
	ParticipationRecordId int64  `json:"participationRecordId"`
	RequestId             string `json:"requestId"`
	ResultType            string `json:"resultType"`
	ResultStatus          string `json:"resultStatus"`
	ResultStatusText      string `json:"resultStatusText"`
	FailureReason         string `json:"failureReason"`
	CreateTime            string `json:"createTime"`
}

type DigitalCardAssetMintTaskSummary struct {
	TaskId             int64    `json:"taskId"`
	RequestId          string   `json:"requestId"`
	TraceId            string   `json:"traceId"`
	TaskStatus         string   `json:"taskStatus"`
	TaskStatusText     string   `json:"taskStatusText"`
	MintStatus         string   `json:"mintStatus"`
	MintStatusText     string   `json:"mintStatusText"`
	ChainStatus        string   `json:"chainStatus"`
	ChainStatusText    string   `json:"chainStatusText"`
	ChainTxId          string   `json:"chainTxId"`
	LastReceiptSummary string   `json:"lastReceiptSummary"`
	AvailableActions   []string `json:"availableActions"`
	ChainType          string   `json:"chainType"`
}

type DigitalCardAssetLogItem struct {
	OperationType string `json:"operationType"`
	OperatorType  string `json:"operatorType"`
	FromStatus    string `json:"fromStatus"`
	ToStatus      string `json:"toStatus"`
	ReasonText    string `json:"reasonText"`
	TraceId       string `json:"traceId"`
	PayloadJson   string `json:"payloadJson"`
	CreateTime    string `json:"createTime"`
}

type DigitalCardAssetDetailData struct {
	Item                  DigitalCardAssetItem                 `json:"item"`
	ParticipationSummary  DigitalCardAssetParticipationSummary `json:"participationSummary"`
	MintTaskSummary       DigitalCardAssetMintTaskSummary      `json:"mintTaskSummary"`
	TraceId               string                               `json:"traceId"`
	RequestId             string                               `json:"requestId"`
	RuleSnapshotJson      string                               `json:"ruleSnapshotJson"`
	Logs                  []DigitalCardAssetLogItem            `json:"logs"`
	AvailableAssetActions []string                             `json:"availableAssetActions"`
}

type QueryDigitalCardAssetDetailResp struct {
	Code    string                     `json:"code"`
	Message string                     `json:"message"`
	Data    DigitalCardAssetDetailData `json:"data"`
	Success bool                       `json:"success"`
}

type QueryDigitalCardAssetLogsReq struct {
	AssetInstanceId int64 `form:"assetInstanceId"`
}

type QueryDigitalCardAssetLogsResp struct {
	Code    string                    `json:"code"`
	Message string                    `json:"message"`
	Logs    []DigitalCardAssetLogItem `json:"logs"`
	Success bool                      `json:"success"`
}

// Story 10.7 Task 9.2 — 后台合规审计跨资产检索
type QueryDigitalCardTransferLogListReq struct {
	DigitalCardAssetQueryGovernanceScopeReq
	PageSize        int    `form:"pageSize,default=20"`
	Current         int    `form:"current,default=1"`
	AssetInstanceId int64  `form:"assetInstanceId,optional"`
	AssetNo         string `form:"assetNo,optional"`
	OperationType   string `form:"operationType,optional"`
	TraceId         string `form:"traceId,optional"`
	DateFrom        string `form:"dateFrom,optional"`
	DateTo          string `form:"dateTo,optional"`
}

type DigitalCardTransferLogListItem struct {
	Id              int64  `json:"id"`
	AssetInstanceId int64  `json:"assetInstanceId"`
	AssetNo         string `json:"assetNo"`
	TemplateName    string `json:"templateName"`
	OperationType   string `json:"operationType"`
	OperatorType    string `json:"operatorType"`
	FromStatus      string `json:"fromStatus"`
	ToStatus        string `json:"toStatus"`
	ReasonCode      string `json:"reasonCode"`
	ReasonText      string `json:"reasonText"`
	TraceId         string `json:"traceId"`
	PayloadJson     string `json:"payloadJson"`
	CreateTime      string `json:"createTime"`
}

type QueryDigitalCardTransferLogListResp struct {
	Code     string                           `json:"code"`
	Message  string                           `json:"message"`
	Total    int64                            `json:"total"`
	Current  int                              `json:"current"`
	PageSize int                              `json:"pageSize"`
	Data     []DigitalCardTransferLogListItem `json:"data"`
	Success  bool                             `json:"success"`
}

type DigitalCardAssetActionReq struct {
	DigitalCardAssetGovernanceScopeReq
	AssetInstanceId int64  `json:"assetInstanceId"`
	Reason          string `json:"reason"`
}

type DigitalCardAssetActionResp struct {
	Code                 string `json:"code"`
	Message              string `json:"message"`
	AssetInstanceId      int64  `json:"assetInstanceId"`
	DisplayStatus        string `json:"displayStatus"`
	DisplayStatusText    string `json:"displayStatusText"`
	ComplianceStatus     string `json:"complianceStatus"`
	ComplianceStatusText string `json:"complianceStatusText"`
	Success              bool   `json:"success"`
}
