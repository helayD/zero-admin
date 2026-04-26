package types

type DigitalCardPhysicalFulfillmentGovernanceScopeReq struct {
	ScopeType  string `form:"scopeType,optional" json:"scopeType,optional"`
	PlatformId int64  `form:"platformId,optional" json:"platformId,optional"`
	TenantId   int64  `form:"tenantId,optional" json:"tenantId,optional"`
	MerchantId int64  `form:"merchantId,optional" json:"merchantId,optional"`
}

type DigitalCardPhysicalFulfillmentQueryGovernanceScopeReq struct {
	ScopeType  string `form:"scopeType,optional"`
	PlatformId int64  `form:"platformId,optional"`
	TenantId   int64  `form:"tenantId,optional"`
	MerchantId int64  `form:"merchantId,optional"`
}

type QueryDigitalCardPhysicalFulfillmentListReq struct {
	DigitalCardPhysicalFulfillmentQueryGovernanceScopeReq
	PageSize          int    `form:"pageSize,default=20"`
	Current           int    `form:"current,default=1"`
	ActivityId        int64  `form:"activityId,optional"`
	TemplateId        int64  `form:"templateId,optional"`
	MemberId          int64  `form:"memberId,optional"`
	AssetNo           string `form:"assetNo,optional"`
	FulfillmentNo     string `form:"fulfillmentNo,optional"`
	ProductionBatchNo string `form:"productionBatchNo,optional"`
	FulfillmentStatus string `form:"fulfillmentStatus,optional"`
	ProductionStatus  string `form:"productionStatus,optional"`
	ShippingStatus    string `form:"shippingStatus,optional"`
	TrackingNo        string `form:"trackingNo,optional"`
	FailureCode       string `form:"failureCode,optional"`
	StartTime         string `form:"startTime,optional"`
	EndTime           string `form:"endTime,optional"`
}

type DigitalCardPhysicalFulfillmentItem struct {
	FulfillmentId         int64  `json:"fulfillmentId"`
	FulfillmentNo         string `json:"fulfillmentNo"`
	AssetInstanceId       int64  `json:"assetInstanceId"`
	AssetNo               string `json:"assetNo"`
	ActivityId            int64  `json:"activityId"`
	ActivityName          string `json:"activityName"`
	TemplateId            int64  `json:"templateId"`
	TemplateName          string `json:"templateName"`
	MemberId              int64  `json:"memberId"`
	FulfillmentStatus     string `json:"fulfillmentStatus"`
	FulfillmentStatusText string `json:"fulfillmentStatusText"`
	ProductionStatus      string `json:"productionStatus"`
	ProductionStatusText  string `json:"productionStatusText"`
	ShippingStatus        string `json:"shippingStatus"`
	ShippingStatusText    string `json:"shippingStatusText"`
	ProductionBatchNo     string `json:"productionBatchNo"`
	CarrierName           string `json:"carrierName"`
	TrackingNo            string `json:"trackingNo"`
	FailureCode           string `json:"failureCode"`
	FailureReason         string `json:"failureReason"`
	ShippedAt             string `json:"shippedAt"`
	SignedAt              string `json:"signedAt"`
	UpdateTime            string `json:"updateTime"`
}

type QueryDigitalCardPhysicalFulfillmentListData struct {
	List  []*DigitalCardPhysicalFulfillmentItem `json:"list"`
	Total int64                                 `json:"total"`
}

type QueryDigitalCardPhysicalFulfillmentListResp struct {
	Code     string                                      `json:"code"`
	Message  string                                      `json:"message"`
	Data     QueryDigitalCardPhysicalFulfillmentListData `json:"data"`
	Current  int                                         `json:"current"`
	PageSize int                                         `json:"pageSize"`
	Total    int64                                       `json:"total"`
	Success  bool                                        `json:"success"`
}

type QueryDigitalCardPhysicalFulfillmentDetailReq struct {
	FulfillmentId int64 `form:"fulfillmentId"`
}

type DigitalCardPhysicalFulfillmentTimelineItem struct {
	Action     string `json:"action"`
	ActionText string `json:"actionText"`
	StatusText string `json:"statusText"`
	Reason     string `json:"reason"`
	CreateTime string `json:"createTime"`
}

type DigitalCardPhysicalFulfillmentDetailData struct {
	Item           DigitalCardPhysicalFulfillmentItem           `json:"item"`
	ReceiverName   string                                       `json:"receiverName"`
	ReceiverPhone  string                                       `json:"receiverPhone"`
	AddressSummary string                                       `json:"addressSummary"`
	RequestId      string                                       `json:"requestId"`
	TraceId        string                                       `json:"traceId"`
	Timeline       []DigitalCardPhysicalFulfillmentTimelineItem `json:"timeline"`
}

type QueryDigitalCardPhysicalFulfillmentDetailResp struct {
	Code    string                                   `json:"code"`
	Message string                                   `json:"message"`
	Data    DigitalCardPhysicalFulfillmentDetailData `json:"data"`
	Success bool                                     `json:"success"`
}

type EnsureDigitalCardPhysicalFulfillmentReq struct {
	DigitalCardPhysicalFulfillmentGovernanceScopeReq
	AssetInstanceId int64 `json:"assetInstanceId"`
}

type UpdateDigitalCardPhysicalProductionStatusReq struct {
	DigitalCardPhysicalFulfillmentGovernanceScopeReq
	FulfillmentId     int64  `json:"fulfillmentId"`
	ProductionBatchNo string `json:"productionBatchNo,optional"`
	ProductionStatus  string `json:"productionStatus"`
	Reason            string `json:"reason"`
}

type ShipDigitalCardPhysicalFulfillmentReq struct {
	DigitalCardPhysicalFulfillmentGovernanceScopeReq
	FulfillmentId int64  `json:"fulfillmentId"`
	CarrierCode   string `json:"carrierCode"`
	CarrierName   string `json:"carrierName"`
	TrackingNo    string `json:"trackingNo"`
	Reason        string `json:"reason"`
}

type DigitalCardPhysicalFulfillmentExceptionReq struct {
	DigitalCardPhysicalFulfillmentGovernanceScopeReq
	FulfillmentId int64  `json:"fulfillmentId"`
	FailureCode   string `json:"failureCode,optional"`
	Reason        string `json:"reason"`
}

type DigitalCardPhysicalFulfillmentActionResp struct {
	Code                  string `json:"code"`
	Message               string `json:"message"`
	FulfillmentId         int64  `json:"fulfillmentId"`
	AssetInstanceId       int64  `json:"assetInstanceId"`
	FulfillmentNo         string `json:"fulfillmentNo"`
	FulfillmentStatus     string `json:"fulfillmentStatus"`
	FulfillmentStatusText string `json:"fulfillmentStatusText"`
	ProductionStatus      string `json:"productionStatus"`
	ProductionStatusText  string `json:"productionStatusText"`
	ShippingStatus        string `json:"shippingStatus"`
	ShippingStatusText    string `json:"shippingStatusText"`
	BlockedReason         string `json:"blockedReason"`
	BlockedReasonText     string `json:"blockedReasonText"`
	Success               bool   `json:"success"`
}
