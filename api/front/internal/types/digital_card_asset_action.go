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
