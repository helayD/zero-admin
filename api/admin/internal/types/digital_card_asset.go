package types

type DigitalCardAssetGovernanceScopeReq struct {
	ScopeType  string `form:"scopeType,optional" json:"scopeType"`
	PlatformId int64  `form:"platformId,optional" json:"platformId"`
	TenantId   int64  `form:"tenantId,optional" json:"tenantId"`
	MerchantId int64  `form:"merchantId,optional" json:"merchantId"`
}

type QueryDigitalCardAssetListReq struct {
	DigitalCardAssetGovernanceScopeReq
	PageSize         int    `form:"pageSize,default=20" json:"pageSize"`
	Current          int    `form:"current,default=1" json:"current"`
	ActivityId       int64  `form:"activityId,optional" json:"activityId"`
	ActivityName     string `form:"activityName,optional" json:"activityName"`
	MemberId         int64  `form:"memberId,optional" json:"memberId"`
	TemplateId       int64  `form:"templateId,optional" json:"templateId"`
	TemplateName     string `form:"templateName,optional" json:"templateName"`
	AssetNo          string `form:"assetNo,optional" json:"assetNo"`
	TokenId          string `form:"tokenId,optional" json:"tokenId"`
	MintStatus       string `form:"mintStatus,optional" json:"mintStatus"`
	ChainStatus      string `form:"chainStatus,optional" json:"chainStatus"`
	DisplayStatus    string `form:"displayStatus,optional" json:"displayStatus"`
	ComplianceStatus string `form:"complianceStatus,optional" json:"complianceStatus"`
	StartTime        string `form:"startTime,optional" json:"startTime"`
	EndTime          string `form:"endTime,optional" json:"endTime"`
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
	AssetInstanceId int64 `form:"assetInstanceId" json:"assetInstanceId"`
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
	AssetInstanceId int64 `form:"assetInstanceId" json:"assetInstanceId"`
}

type QueryDigitalCardAssetLogsResp struct {
	Code    string                    `json:"code"`
	Message string                    `json:"message"`
	Logs    []DigitalCardAssetLogItem `json:"logs"`
	Success bool                      `json:"success"`
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
