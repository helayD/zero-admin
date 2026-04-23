package antchain

import (
	"github.com/feihua/zero-admin/pkg/chainclient"
)

var ErrReceiptNotFound = chainclient.ErrReceiptNotFound

type Config struct {
	Endpoint        string
	ReceiptEndpoint string
	AppID           string
	AccessKey       string
	Secret          string
	TimeoutSeconds  int64
	Enabled         bool
}

type MintTokenRequest = chainclient.MintTokenRequest
type QueryMintTokenRequest = chainclient.QueryMintTokenRequest
type MintTokenResponse = chainclient.MintTokenResponse
