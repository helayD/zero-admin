package types

type ResolveDigitalCardTransferRecipientReq struct {
	AssetInstanceId int64  `json:"assetInstanceId"`
	RecipientMobile string `json:"recipientMobile"`
}

type DigitalCardTransferRecipientData struct {
	RecipientStatus         string `json:"recipientStatus"`
	RecipientStatusText     string `json:"recipientStatusText"`
	RecipientMemberId       int64  `json:"recipientMemberId"`
	RecipientNicknameMasked string `json:"recipientNicknameMasked"`
	RecipientMobileMasked   string `json:"recipientMobileMasked"`
	RegisterUrl             string `json:"registerUrl"`
	CanTransfer             bool   `json:"canTransfer"`
	ActionHint              string `json:"actionHint"`
}

type ResolveDigitalCardTransferRecipientResp struct {
	Code    string                           `json:"code"`
	Message string                           `json:"message"`
	Data    DigitalCardTransferRecipientData `json:"data"`
}

type TransferDigitalCardAssetReq struct {
	AssetInstanceId int64  `json:"assetInstanceId"`
	RecipientMobile string `json:"recipientMobile"`
	RequestId       string `json:"requestId,optional"`
}

type TransferDigitalCardAssetData struct {
	AssetInstanceId         int64  `json:"assetInstanceId"`
	FromMemberId            int64  `json:"fromMemberId"`
	ToMemberId              int64  `json:"toMemberId"`
	RecipientNicknameMasked string `json:"recipientNicknameMasked"`
	RecipientMobileMasked   string `json:"recipientMobileMasked"`
	TransferStatus          string `json:"transferStatus"`
	TransferStatusText      string `json:"transferStatusText"`
}

type TransferDigitalCardAssetResp struct {
	Code    string                       `json:"code"`
	Message string                       `json:"message"`
	Data    TransferDigitalCardAssetData `json:"data"`
}

type RequestDigitalCardWithdrawReq struct {
	AssetInstanceId int64  `json:"assetInstanceId"`
	Reason          string `json:"reason,optional"`
	RequestId       string `json:"requestId,optional"`
}

type RequestDigitalCardWithdrawData struct {
	AssetInstanceId int64  `json:"assetInstanceId"`
	WithdrawStatus  string `json:"withdrawStatus"`
	WithdrawText    string `json:"withdrawText"`
}

type RequestDigitalCardWithdrawResp struct {
	Code    string                         `json:"code"`
	Message string                         `json:"message"`
	Data    RequestDigitalCardWithdrawData `json:"data"`
}

type CreateRedemptionOrderReq struct {
	CardInstanceId  int64  `json:"cardInstanceId"`
	ReceiverName    string `json:"receiverName"`
	ReceiverPhone   string `json:"receiverPhone"`
	ReceiverAddress string `json:"receiverAddress"`
	PlatformId      int64  `json:"platformId,optional"`
	TenantId        int64  `json:"tenantId,optional"`
	MerchantId      int64  `json:"merchantId,optional"`
	TraceId         string `json:"traceId,optional"`
	RequestId       string `json:"requestId,optional"`
}

type CreateRedemptionOrderResp struct {
	Code    int64                `json:"code"`
	Message string               `json:"message"`
	Data    *RedemptionOrderData `json:"data"`
}

type QueryRedemptionOrderReq struct {
	OrderId        int64 `json:"orderId,optional"`
	CardInstanceId int64 `json:"cardInstanceId,optional"`
}

type QueryRedemptionOrderResp struct {
	Code    int64                `json:"code"`
	Message string               `json:"message"`
	Data    *RedemptionOrderData `json:"data"`
}

type RedemptionOrderData struct {
	Id              int64  `json:"id"`
	OrderNo         string `json:"orderNo"`
	CardInstanceId  int64  `json:"cardInstanceId"`
	HolderId        int64  `json:"holderId"`
	ReceiverName    string `json:"receiverName"`
	ReceiverPhone   string `json:"receiverPhone"`
	ReceiverAddress string `json:"receiverAddress"`
	Status          string `json:"status"`
	ShippedAt       string `json:"shippedAt"`
	DeliveredAt     string `json:"deliveredAt"`
	CancelReason    string `json:"cancelReason"`
	OmsOrderId      int64  `json:"omsOrderId"`
	PlatformId      int64  `json:"platformId"`
	TenantId        int64  `json:"tenantId"`
	MerchantId      int64  `json:"merchantId"`
	CreateTime      string `json:"createTime"`
	UpdateTime      string `json:"updateTime"`
}

type GenerateShareLinkReq struct {
	CardInstanceId int64  `json:"cardInstanceId"`
	ExpireHours    int32  `json:"expireHours,optional"`
	MaxClaims      int32  `json:"maxClaims,optional"`
	Domain         string `json:"domain"`
	PlatformId     int64  `json:"platformId,optional"`
	TenantId       int64  `json:"tenantId,optional"`
	MerchantId     int64  `json:"merchantId,optional"`
	TraceId        string `json:"traceId,optional"`
	RequestId      string `json:"requestId,optional"`
}

type GenerateShareLinkResp struct {
	Code    int64         `json:"code"`
	Message string        `json:"message"`
	Data    ShareLinkData `json:"data"`
}

type ShareLinkData struct {
	Token     string `json:"token"`
	ShareLink string `json:"shareLink"`
	ExpireAt  string `json:"expireAt"`
	MaxClaims int32  `json:"maxClaims"`
}

type ClaimDigitalCardReq struct {
	Token     string `json:"token"`
	TraceId   string `json:"traceId,optional"`
	RequestId string `json:"requestId,optional"`
}

type ClaimDigitalCardResp struct {
	Code    int64              `json:"code"`
	Message string             `json:"message"`
	Data    ClaimDigitalCardData `json:"data"`
}

type ClaimDigitalCardData struct {
	CardInstanceId int64  `json:"cardInstanceId"`
	AssetNo        string `json:"assetNo"`
	Status         string `json:"status"`
}

type GenerateClaimTokenReq struct {
	CardInstanceId int64  `json:"cardInstanceId"`
	IssuerId       int64  `json:"issuerId"`
	ExpireHours    int32  `json:"expireHours"`
	MaxClaims      int32  `json:"maxClaims"`
	PlatformId     int64  `json:"platformId"`
	TenantId       int64  `json:"tenantId"`
	MerchantId     int64  `json:"merchantId"`
	TraceId        string `json:"traceId"`
	RequestId      string `json:"requestId"`
}

type GenerateClaimTokenResp struct {
	Token *ClaimTokenData `json:"token"`
}

type ClaimTokenData struct {
	Id             int64  `json:"id"`
	Token          string `json:"token"`
	CardInstanceId int64  `json:"cardInstanceId"`
	IssuerId       int64  `json:"issuerId"`
	IssuerType     string `json:"issuerType"`
	ExpireAt       string `json:"expireAt"`
	MaxClaims      int32  `json:"maxClaims"`
	ClaimedCount   int32  `json:"claimedCount"`
	Status         string `json:"status"`
	ClaimedBy      int64  `json:"claimedBy"`
	ClaimedAt      string `json:"claimedAt"`
	PlatformId     int64  `json:"platformId"`
	TenantId       int64  `json:"tenantId"`
	MerchantId     int64  `json:"merchantId"`
	CreateTime     string `json:"createTime"`
	UpdateTime     string `json:"updateTime"`
}

type ValidateClaimTokenReq struct {
	Token string `json:"token"`
}

type ValidateClaimTokenResp struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Data    ValidateClaimTokenData `json:"data"`
}

type ValidateClaimTokenData struct {
	Valid         bool           `json:"valid"`
	FailureReason string         `json:"failureReason"`
	Token         *ClaimTokenData `json:"token"`
}

type ConsumeClaimTokenReq struct {
	Token     string `json:"token"`
	ClaimedBy int64  `json:"claimedBy"`
	TraceId   string `json:"traceId"`
	RequestId string `json:"requestId"`
}

type ConsumeClaimTokenResp struct {
	Success       bool              `json:"success"`
	Token         *ClaimTokenData   `json:"token"`
	CardInstance  *CardInstanceData `json:"cardInstance"`
	FailureReason string            `json:"failureReason"`
}

type CardInstanceData struct {
	Id              int64  `json:"id"`
	AssetNo         string `json:"assetNo"`
	AssetStatus     string `json:"assetStatus"`
	AssetStatusText string `json:"assetStatusText"`
	MintStatus      string `json:"mintStatus"`
	TemplateId      int64  `json:"templateId"`
	MemberId        int64  `json:"memberId"`
	IssuedAt        string `json:"issuedAt"`
}
