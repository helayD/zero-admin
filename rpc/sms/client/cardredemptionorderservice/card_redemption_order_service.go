package cardredemptionorderservice

import (
	"context"

	"github.com/feihua/zero-admin/rpc/sms/smsclient"

	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

type (
	CreateRedemptionOrderReq          = smsclient.CreateRedemptionOrderReq
	CreateRedemptionOrderResp         = smsclient.CreateRedemptionOrderResp
	QueryRedemptionOrderReq           = smsclient.QueryRedemptionOrderReq
	QueryRedemptionOrderResp          = smsclient.QueryRedemptionOrderResp
	CancelRedemptionOrderReq          = smsclient.CancelRedemptionOrderReq
	CancelRedemptionOrderResp         = smsclient.CancelRedemptionOrderResp
	UpdateRedemptionOrderStatusReq    = smsclient.UpdateRedemptionOrderStatusReq
	UpdateRedemptionOrderStatusResp   = smsclient.UpdateRedemptionOrderStatusResp
	RedemptionOrderData               = smsclient.RedemptionOrderData

	CardRedemptionOrderService interface {
		CreateRedemptionOrder(ctx context.Context, in *CreateRedemptionOrderReq, opts ...grpc.CallOption) (*CreateRedemptionOrderResp, error)
		QueryRedemptionOrder(ctx context.Context, in *QueryRedemptionOrderReq, opts ...grpc.CallOption) (*QueryRedemptionOrderResp, error)
		CancelRedemptionOrder(ctx context.Context, in *CancelRedemptionOrderReq, opts ...grpc.CallOption) (*CancelRedemptionOrderResp, error)
		UpdateRedemptionOrderStatus(ctx context.Context, in *UpdateRedemptionOrderStatusReq, opts ...grpc.CallOption) (*UpdateRedemptionOrderStatusResp, error)
	}

	defaultCardRedemptionOrderService struct {
		cli zrpc.Client
	}
)

func NewCardRedemptionOrderService(cli zrpc.Client) CardRedemptionOrderService {
	return &defaultCardRedemptionOrderService{
		cli: cli,
	}
}

func (m *defaultCardRedemptionOrderService) CreateRedemptionOrder(ctx context.Context, in *CreateRedemptionOrderReq, opts ...grpc.CallOption) (*CreateRedemptionOrderResp, error) {
	client := smsclient.NewCardRedemptionOrderServiceClient(m.cli.Conn())
	return client.CreateRedemptionOrder(ctx, in, opts...)
}

func (m *defaultCardRedemptionOrderService) QueryRedemptionOrder(ctx context.Context, in *QueryRedemptionOrderReq, opts ...grpc.CallOption) (*QueryRedemptionOrderResp, error) {
	client := smsclient.NewCardRedemptionOrderServiceClient(m.cli.Conn())
	return client.QueryRedemptionOrder(ctx, in, opts...)
}

func (m *defaultCardRedemptionOrderService) CancelRedemptionOrder(ctx context.Context, in *CancelRedemptionOrderReq, opts ...grpc.CallOption) (*CancelRedemptionOrderResp, error) {
	client := smsclient.NewCardRedemptionOrderServiceClient(m.cli.Conn())
	return client.CancelRedemptionOrder(ctx, in, opts...)
}

func (m *defaultCardRedemptionOrderService) UpdateRedemptionOrderStatus(ctx context.Context, in *UpdateRedemptionOrderStatusReq, opts ...grpc.CallOption) (*UpdateRedemptionOrderStatusResp, error) {
	client := smsclient.NewCardRedemptionOrderServiceClient(m.cli.Conn())
	return client.UpdateRedemptionOrderStatus(ctx, in, opts...)
}
