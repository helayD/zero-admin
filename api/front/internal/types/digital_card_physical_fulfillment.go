package types

type QueryMyPhysicalFulfillmentDetailReq struct {
	AssetInstanceId int64 `form:"assetInstanceId"`
}

type ConfirmPhysicalFulfillmentAddressReq struct {
	AssetInstanceId int64 `json:"assetInstanceId"`
	AddressId       int64 `json:"addressId"`
}

type ConfirmPhysicalCardReceiptReq struct {
	FulfillmentId int64  `json:"fulfillmentId"`
	Reason        string `json:"reason,optional"`
}

type ConfirmPhysicalFulfillmentShippingFeeReq struct {
	FulfillmentId   int64  `json:"fulfillmentId,optional"`
	AssetInstanceId int64  `json:"assetInstanceId"`
	PayAmount       int64  `json:"payAmount,optional"`
	PayChannel      string `json:"payChannel,optional"`
	PaymentNo       string `json:"paymentNo,optional"`
}

type PhysicalFulfillmentTimelineItem struct {
	Action     string `json:"action"`
	ActionText string `json:"actionText"`
	StatusText string `json:"statusText"`
	Reason     string `json:"reason"`
	CreateTime string `json:"createTime"`
}

type PhysicalFulfillmentDetailData struct {
	FulfillmentId         int64                             `json:"fulfillmentId"`
	FulfillmentNo         string                            `json:"fulfillmentNo"`
	AssetInstanceId       int64                             `json:"assetInstanceId"`
	AssetNo               string                            `json:"assetNo"`
	TemplateName          string                            `json:"templateName"`
	ActivityName          string                            `json:"activityName"`
	ObtainedAt            string                            `json:"obtainedAt"`
	MintStatusText        string                            `json:"mintStatusText"`
	FulfillmentStatus     string                            `json:"fulfillmentStatus"`
	FulfillmentStatusText string                            `json:"fulfillmentStatusText"`
	ProductionStatusText  string                            `json:"productionStatusText"`
	ShippingStatusText    string                            `json:"shippingStatusText"`
	ShippingFeeStatus     string                            `json:"shippingFeeStatus"`
	ShippingFeeStatusText string                            `json:"shippingFeeStatusText"`
	ShippingFeeAmount     int64                             `json:"shippingFeeAmount"`
	ReceiverNameMasked    string                            `json:"receiverNameMasked"`
	ReceiverPhoneMasked   string                            `json:"receiverPhoneMasked"`
	AddressSummary        string                            `json:"addressSummary"`
	CarrierName           string                            `json:"carrierName"`
	TrackingNo            string                            `json:"trackingNo"`
	ComplianceTipSummary  string                            `json:"complianceTipSummary"`
	BlockedReason         string                            `json:"blockedReason"`
	BlockedReasonText     string                            `json:"blockedReasonText"`
	Timeline              []PhysicalFulfillmentTimelineItem `json:"timeline"`
}

type QueryMyPhysicalFulfillmentDetailResp struct {
	Code    string                        `json:"code"`
	Message string                        `json:"message"`
	Data    PhysicalFulfillmentDetailData `json:"data"`
}

type ConfirmPhysicalFulfillmentAddressResp struct {
	Code    string                        `json:"code"`
	Message string                        `json:"message"`
	Data    PhysicalFulfillmentDetailData `json:"data"`
}

type ConfirmPhysicalCardReceiptResp struct {
	Code    string                        `json:"code"`
	Message string                        `json:"message"`
	Data    PhysicalFulfillmentDetailData `json:"data"`
}

type ConfirmPhysicalFulfillmentShippingFeeResp struct {
	Code    string                        `json:"code"`
	Message string                        `json:"message"`
	Data    PhysicalFulfillmentDetailData `json:"data"`
}
