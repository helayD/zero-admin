package server

import (
	"context"

	"github.com/feihua/zero-admin/rpc/sms/internal/logic/cardredemptionorderservice"
	"github.com/feihua/zero-admin/rpc/sms/internal/svc"
	"github.com/feihua/zero-admin/rpc/sms/smsclient"
)

type CardRedemptionOrderServiceServer struct {
	svcCtx *svc.ServiceContext
	smsclient.UnimplementedCardRedemptionOrderServiceServer
}

func NewCardRedemptionOrderServiceServer(svcCtx *svc.ServiceContext) *CardRedemptionOrderServiceServer {
	return &CardRedemptionOrderServiceServer{
		svcCtx: svcCtx,
	}
}

func (s *CardRedemptionOrderServiceServer) CreateRedemptionOrder(ctx context.Context, in *smsclient.CreateRedemptionOrderReq) (*smsclient.CreateRedemptionOrderResp, error) {
	l := cardredemptionorderservicelogic.NewCreateRedemptionOrderLogic(ctx, s.svcCtx)
	return l.CreateRedemptionOrder(in)
}

func (s *CardRedemptionOrderServiceServer) QueryRedemptionOrder(ctx context.Context, in *smsclient.QueryRedemptionOrderReq) (*smsclient.QueryRedemptionOrderResp, error) {
	l := cardredemptionorderservicelogic.NewQueryRedemptionOrderLogic(ctx, s.svcCtx)
	return l.QueryRedemptionOrder(in)
}

func (s *CardRedemptionOrderServiceServer) CancelRedemptionOrder(ctx context.Context, in *smsclient.CancelRedemptionOrderReq) (*smsclient.CancelRedemptionOrderResp, error) {
	l := cardredemptionorderservicelogic.NewCancelRedemptionOrderLogic(ctx, s.svcCtx)
	return l.CancelRedemptionOrder(in)
}

func (s *CardRedemptionOrderServiceServer) UpdateRedemptionOrderStatus(ctx context.Context, in *smsclient.UpdateRedemptionOrderStatusReq) (*smsclient.UpdateRedemptionOrderStatusResp, error) {
	l := cardredemptionorderservicelogic.NewUpdateRedemptionOrderStatusLogic(ctx, s.svcCtx)
	return l.UpdateRedemptionOrderStatus(in)
}
