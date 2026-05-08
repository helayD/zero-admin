package cardclaimtokenservice

import (
	"context"

	"github.com/feihua/zero-admin/rpc/sms/smsclient"

	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

type (
	GenerateClaimTokenReq  = smsclient.GenerateClaimTokenReq
	GenerateClaimTokenResp = smsclient.GenerateClaimTokenResp
	ValidateClaimTokenReq  = smsclient.ValidateClaimTokenReq
	ValidateClaimTokenResp = smsclient.ValidateClaimTokenResp
	ConsumeClaimTokenReq   = smsclient.ConsumeClaimTokenReq
	ConsumeClaimTokenResp  = smsclient.ConsumeClaimTokenResp
	RevokeClaimTokenReq    = smsclient.RevokeClaimTokenReq
	RevokeClaimTokenResp   = smsclient.RevokeClaimTokenResp
	ClaimTokenData         = smsclient.ClaimTokenData

	CardClaimTokenService interface {
		GenerateClaimToken(ctx context.Context, in *GenerateClaimTokenReq, opts ...grpc.CallOption) (*GenerateClaimTokenResp, error)
		ValidateClaimToken(ctx context.Context, in *ValidateClaimTokenReq, opts ...grpc.CallOption) (*ValidateClaimTokenResp, error)
		ConsumeClaimToken(ctx context.Context, in *ConsumeClaimTokenReq, opts ...grpc.CallOption) (*ConsumeClaimTokenResp, error)
		RevokeClaimToken(ctx context.Context, in *RevokeClaimTokenReq, opts ...grpc.CallOption) (*RevokeClaimTokenResp, error)
	}

	defaultCardClaimTokenService struct {
		cli zrpc.Client
	}
)

func NewCardClaimTokenService(cli zrpc.Client) CardClaimTokenService {
	return &defaultCardClaimTokenService{
		cli: cli,
	}
}

func (m *defaultCardClaimTokenService) GenerateClaimToken(ctx context.Context, in *GenerateClaimTokenReq, opts ...grpc.CallOption) (*GenerateClaimTokenResp, error) {
	client := smsclient.NewCardClaimTokenServiceClient(m.cli.Conn())
	return client.GenerateClaimToken(ctx, in, opts...)
}

func (m *defaultCardClaimTokenService) ValidateClaimToken(ctx context.Context, in *ValidateClaimTokenReq, opts ...grpc.CallOption) (*ValidateClaimTokenResp, error) {
	client := smsclient.NewCardClaimTokenServiceClient(m.cli.Conn())
	return client.ValidateClaimToken(ctx, in, opts...)
}

func (m *defaultCardClaimTokenService) ConsumeClaimToken(ctx context.Context, in *ConsumeClaimTokenReq, opts ...grpc.CallOption) (*ConsumeClaimTokenResp, error) {
	client := smsclient.NewCardClaimTokenServiceClient(m.cli.Conn())
	return client.ConsumeClaimToken(ctx, in, opts...)
}

func (m *defaultCardClaimTokenService) RevokeClaimToken(ctx context.Context, in *RevokeClaimTokenReq, opts ...grpc.CallOption) (*RevokeClaimTokenResp, error) {
	client := smsclient.NewCardClaimTokenServiceClient(m.cli.Conn())
	return client.RevokeClaimToken(ctx, in, opts...)
}
