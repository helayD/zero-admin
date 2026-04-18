package antchain

import "time"

type Config struct {
	Endpoint       string
	AppID          string
	AccessKey      string
	Secret         string
	TimeoutSeconds int64
	Enabled        bool
}

type MintTokenRequest struct {
	IdempotencyKey string
	TaskID         int64
	AssetInstanceID int64
	ActivityID     int64
	TemplateID     int64
	MemberID       int64
	RequestID      string
	TraceID        string
	AssetNo        string
	ScopeType      string
	PlatformID     int64
	TenantID       int64
	MerchantID     int64
}

type MintTokenResponse struct {
	TokenID        string
	ChainTxID      string
	ChainStatus    string
	ReceiptSummary string
	ReceiptJSON    string
	ConfirmedAt    time.Time
}

