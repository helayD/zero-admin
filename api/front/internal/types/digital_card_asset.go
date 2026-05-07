package types

type QueryMyDigitalCardAssetListReq struct {
	PageNum  int32 `form:"pageNum,default=1"`
	PageSize int32 `form:"pageSize,default=20"`
}

type DigitalCardAssetItem struct {
	AssetInstanceId       int64  `json:"assetInstanceId"`
	AssetNo               string `json:"assetNo"`
	TemplateId            int64  `json:"templateId"`
	TemplateName          string `json:"templateName"`
	CardFaceImage         string `json:"cardFaceImage"`
	ActivityId            int64  `json:"activityId"`
	ActivityName          string `json:"activityName"`
	SourceType            string `json:"sourceType"`
	SourceDisplayName     string `json:"sourceDisplayName"`
	Rarity                string `json:"rarity"`
	ObtainedAt            string `json:"obtainedAt"`
	MintStatus            string `json:"mintStatus"`
	MintStatusText        string `json:"mintStatusText"`
	DisplayStatus         string `json:"displayStatus"`
	DisplayStatusText     string `json:"displayStatusText"`
	ComplianceStatus      string `json:"complianceStatus"`
	ComplianceStatusText  string `json:"complianceStatusText"`
	TokenStatusText       string `json:"tokenStatusText"`
	ComplianceRuleSummary string `json:"complianceRuleSummary"`
}

type QueryMyDigitalCardAssetListData struct {
	Total int64                  `json:"total"`
	List  []DigitalCardAssetItem `json:"list"`
}

type QueryMyDigitalCardAssetListResp struct {
	Code    string                          `json:"code"`
	Message string                          `json:"message"`
	Data    QueryMyDigitalCardAssetListData `json:"data"`
}

type QueryMyDigitalCardAssetDetailReq struct {
	AssetInstanceId int64 `form:"assetInstanceId"`
}

type DigitalCardAssetTimelineItem struct {
	OperationType string `json:"operationType"`
	OperationText string `json:"operationText"`
	StatusText    string `json:"statusText"`
	ReasonText    string `json:"reasonText"`
	CreateTime    string `json:"createTime"`
}

type DigitalCardAssetDrawSummary struct {
	ParticipationRecordId int64  `json:"participationRecordId"`
	ResultType            string `json:"resultType"`
	ResultStatus          string `json:"resultStatus"`
	ResultStatusText      string `json:"resultStatusText"`
	FailureReason         string `json:"failureReason"`
	CreateTime            string `json:"createTime"`
}

type DigitalCardAssetDetailData struct {
	Item                DigitalCardAssetItem           `json:"item"`
	LatestStatusSummary string                         `json:"latestStatusSummary"`
	RestrictionReason   string                         `json:"restrictionReason"`
	DrawSummary         DigitalCardAssetDrawSummary    `json:"drawSummary"`
	Timeline            []DigitalCardAssetTimelineItem `json:"timeline"`
}

type QueryMyDigitalCardAssetDetailResp struct {
	Code    string                     `json:"code"`
	Message string                     `json:"message"`
	Data    DigitalCardAssetDetailData `json:"data"`
}
