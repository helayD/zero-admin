package chainclient

import "context"

type ChainClient interface {
	MintToken(ctx context.Context, req *MintTokenRequest) (*MintTokenResponse, error)
	QueryMintToken(ctx context.Context, req *QueryMintTokenRequest) (*MintTokenResponse, error)
	ChainType() string
}
