package productfulfillmentruleservice

import (
	"context"

	"github.com/feihua/zero-admin/rpc/sms/smsclient"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

type (
	AddProductFulfillmentRuleReq         = smsclient.AddProductFulfillmentRuleReq
	AddProductFulfillmentRuleResp        = smsclient.AddProductFulfillmentRuleResp
	UpdateProductFulfillmentRuleReq      = smsclient.UpdateProductFulfillmentRuleReq
	UpdateProductFulfillmentRuleResp     = smsclient.UpdateProductFulfillmentRuleResp
	QueryProductFulfillmentRuleListReq   = smsclient.QueryProductFulfillmentRuleListReq
	QueryProductFulfillmentRuleListResp  = smsclient.QueryProductFulfillmentRuleListResp
	QueryProductFulfillmentRuleDetailReq = smsclient.QueryProductFulfillmentRuleDetailReq
	QueryProductFulfillmentRuleDetailResp = smsclient.QueryProductFulfillmentRuleDetailResp
	UpdateProductFulfillmentRuleStatusReq  = smsclient.UpdateProductFulfillmentRuleStatusReq
	UpdateProductFulfillmentRuleStatusResp = smsclient.UpdateProductFulfillmentRuleStatusResp
	DeleteProductFulfillmentRuleReq      = smsclient.DeleteProductFulfillmentRuleReq
	DeleteProductFulfillmentRuleResp     = smsclient.DeleteProductFulfillmentRuleResp
	CheckProductFulfillmentRuleBindingReq  = smsclient.CheckProductFulfillmentRuleBindingReq
	CheckProductFulfillmentRuleBindingResp = smsclient.CheckProductFulfillmentRuleBindingResp
	ProductFulfillmentRuleData           = smsclient.ProductFulfillmentRuleData

	ProductFulfillmentRuleService interface {
		AddProductFulfillmentRule(ctx context.Context, in *AddProductFulfillmentRuleReq, opts ...grpc.CallOption) (*AddProductFulfillmentRuleResp, error)
		UpdateProductFulfillmentRule(ctx context.Context, in *UpdateProductFulfillmentRuleReq, opts ...grpc.CallOption) (*UpdateProductFulfillmentRuleResp, error)
		QueryProductFulfillmentRuleList(ctx context.Context, in *QueryProductFulfillmentRuleListReq, opts ...grpc.CallOption) (*QueryProductFulfillmentRuleListResp, error)
		QueryProductFulfillmentRuleDetail(ctx context.Context, in *QueryProductFulfillmentRuleDetailReq, opts ...grpc.CallOption) (*QueryProductFulfillmentRuleDetailResp, error)
		UpdateProductFulfillmentRuleStatus(ctx context.Context, in *UpdateProductFulfillmentRuleStatusReq, opts ...grpc.CallOption) (*UpdateProductFulfillmentRuleStatusResp, error)
		DeleteProductFulfillmentRule(ctx context.Context, in *DeleteProductFulfillmentRuleReq, opts ...grpc.CallOption) (*DeleteProductFulfillmentRuleResp, error)
		CheckProductFulfillmentRuleBinding(ctx context.Context, in *CheckProductFulfillmentRuleBindingReq, opts ...grpc.CallOption) (*CheckProductFulfillmentRuleBindingResp, error)
	}

	defaultProductFulfillmentRuleService struct {
		cli zrpc.Client
	}
)

func NewProductFulfillmentRuleService(cli zrpc.Client) ProductFulfillmentRuleService {
	return &defaultProductFulfillmentRuleService{
		cli: cli,
	}
}

func (m *defaultProductFulfillmentRuleService) AddProductFulfillmentRule(ctx context.Context, in *AddProductFulfillmentRuleReq, opts ...grpc.CallOption) (*AddProductFulfillmentRuleResp, error) {
	client := smsclient.NewProductFulfillmentRuleServiceClient(m.cli.Conn())
	return client.AddProductFulfillmentRule(ctx, in, opts...)
}

func (m *defaultProductFulfillmentRuleService) UpdateProductFulfillmentRule(ctx context.Context, in *UpdateProductFulfillmentRuleReq, opts ...grpc.CallOption) (*UpdateProductFulfillmentRuleResp, error) {
	client := smsclient.NewProductFulfillmentRuleServiceClient(m.cli.Conn())
	return client.UpdateProductFulfillmentRule(ctx, in, opts...)
}

func (m *defaultProductFulfillmentRuleService) QueryProductFulfillmentRuleList(ctx context.Context, in *QueryProductFulfillmentRuleListReq, opts ...grpc.CallOption) (*QueryProductFulfillmentRuleListResp, error) {
	client := smsclient.NewProductFulfillmentRuleServiceClient(m.cli.Conn())
	return client.QueryProductFulfillmentRuleList(ctx, in, opts...)
}

func (m *defaultProductFulfillmentRuleService) QueryProductFulfillmentRuleDetail(ctx context.Context, in *QueryProductFulfillmentRuleDetailReq, opts ...grpc.CallOption) (*QueryProductFulfillmentRuleDetailResp, error) {
	client := smsclient.NewProductFulfillmentRuleServiceClient(m.cli.Conn())
	return client.QueryProductFulfillmentRuleDetail(ctx, in, opts...)
}

func (m *defaultProductFulfillmentRuleService) UpdateProductFulfillmentRuleStatus(ctx context.Context, in *UpdateProductFulfillmentRuleStatusReq, opts ...grpc.CallOption) (*UpdateProductFulfillmentRuleStatusResp, error) {
	client := smsclient.NewProductFulfillmentRuleServiceClient(m.cli.Conn())
	return client.UpdateProductFulfillmentRuleStatus(ctx, in, opts...)
}

func (m *defaultProductFulfillmentRuleService) DeleteProductFulfillmentRule(ctx context.Context, in *DeleteProductFulfillmentRuleReq, opts ...grpc.CallOption) (*DeleteProductFulfillmentRuleResp, error) {
	client := smsclient.NewProductFulfillmentRuleServiceClient(m.cli.Conn())
	return client.DeleteProductFulfillmentRule(ctx, in, opts...)
}

func (m *defaultProductFulfillmentRuleService) CheckProductFulfillmentRuleBinding(ctx context.Context, in *CheckProductFulfillmentRuleBindingReq, opts ...grpc.CallOption) (*CheckProductFulfillmentRuleBindingResp, error) {
	client := smsclient.NewProductFulfillmentRuleServiceClient(m.cli.Conn())
	return client.CheckProductFulfillmentRuleBinding(ctx, in, opts...)
}
