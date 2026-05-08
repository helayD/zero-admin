package server

import (
	"context"

	"github.com/feihua/zero-admin/rpc/sms/internal/logic/cardclaimtokenservice"
	"github.com/feihua/zero-admin/rpc/sms/internal/svc"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"
)

type CardClaimTokenServiceServer struct {
	svcCtx *svc.ServiceContext
	smsclient.UnimplementedCardClaimTokenServiceServer
}

func NewCardClaimTokenServiceServer(svcCtx *svc.ServiceContext) *CardClaimTokenServiceServer {
	return &CardClaimTokenServiceServer{
		svcCtx: svcCtx,
	}
}

func (s *CardClaimTokenServiceServer) GenerateClaimToken(ctx context.Context, in *smsclient.GenerateClaimTokenReq) (*smsclient.GenerateClaimTokenResp, error) {
	l := cardclaimtokenservicelogic.NewGenerateClaimTokenLogic(ctx, s.svcCtx)
	return l.GenerateClaimToken(in)
}

func (s *CardClaimTokenServiceServer) ValidateClaimToken(ctx context.Context, in *smsclient.ValidateClaimTokenReq) (*smsclient.ValidateClaimTokenResp, error) {
	l := cardclaimtokenservicelogic.NewValidateClaimTokenLogic(ctx, s.svcCtx)
	return l.ValidateClaimToken(in)
}

func (s *CardClaimTokenServiceServer) ConsumeClaimToken(ctx context.Context, in *smsclient.ConsumeClaimTokenReq) (*smsclient.ConsumeClaimTokenResp, error) {
	l := cardclaimtokenservicelogic.NewConsumeClaimTokenLogic(ctx, s.svcCtx)
	return l.ConsumeClaimToken(in)
}

func (s *CardClaimTokenServiceServer) RevokeClaimToken(ctx context.Context, in *smsclient.RevokeClaimTokenReq) (*smsclient.RevokeClaimTokenResp, error) {
	l := cardclaimtokenservicelogic.NewRevokeClaimTokenLogic(ctx, s.svcCtx)
	return l.RevokeClaimToken(in)
}
