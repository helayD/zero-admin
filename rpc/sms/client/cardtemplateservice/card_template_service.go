package cardtemplateservice

import (
	"context"

	"github.com/feihua/zero-admin/rpc/sms/smsclient"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

type (
	AddCardTemplateReq           = smsclient.AddCardTemplateReq
	AddCardTemplateResp          = smsclient.AddCardTemplateResp
	UpdateCardTemplateReq        = smsclient.UpdateCardTemplateReq
	UpdateCardTemplateResp       = smsclient.UpdateCardTemplateResp
	QueryCardTemplateListReq     = smsclient.QueryCardTemplateListReq
	QueryCardTemplateListResp    = smsclient.QueryCardTemplateListResp
	QueryCardTemplateDetailReq   = smsclient.QueryCardTemplateDetailReq
	QueryCardTemplateDetailResp  = smsclient.QueryCardTemplateDetailResp
	UpdateCardTemplateStatusReq  = smsclient.UpdateCardTemplateStatusReq
	UpdateCardTemplateStatusResp = smsclient.UpdateCardTemplateStatusResp
	DeleteCardTemplateReq        = smsclient.DeleteCardTemplateReq
	DeleteCardTemplateResp       = smsclient.DeleteCardTemplateResp
	CheckCardTemplateUsageReq    = smsclient.CheckCardTemplateUsageReq
	CheckCardTemplateUsageResp   = smsclient.CheckCardTemplateUsageResp
	CardTemplateData             = smsclient.CardTemplateData

	CardTemplateService interface {
		AddCardTemplate(ctx context.Context, in *AddCardTemplateReq, opts ...grpc.CallOption) (*AddCardTemplateResp, error)
		UpdateCardTemplate(ctx context.Context, in *UpdateCardTemplateReq, opts ...grpc.CallOption) (*UpdateCardTemplateResp, error)
		QueryCardTemplateList(ctx context.Context, in *QueryCardTemplateListReq, opts ...grpc.CallOption) (*QueryCardTemplateListResp, error)
		QueryCardTemplateDetail(ctx context.Context, in *QueryCardTemplateDetailReq, opts ...grpc.CallOption) (*QueryCardTemplateDetailResp, error)
		UpdateCardTemplateStatus(ctx context.Context, in *UpdateCardTemplateStatusReq, opts ...grpc.CallOption) (*UpdateCardTemplateStatusResp, error)
		DeleteCardTemplate(ctx context.Context, in *DeleteCardTemplateReq, opts ...grpc.CallOption) (*DeleteCardTemplateResp, error)
		CheckCardTemplateUsage(ctx context.Context, in *CheckCardTemplateUsageReq, opts ...grpc.CallOption) (*CheckCardTemplateUsageResp, error)
	}

	defaultCardTemplateService struct {
		cli zrpc.Client
	}
)

func NewCardTemplateService(cli zrpc.Client) CardTemplateService {
	return &defaultCardTemplateService{cli: cli}
}

func (m *defaultCardTemplateService) AddCardTemplate(ctx context.Context, in *AddCardTemplateReq, opts ...grpc.CallOption) (*AddCardTemplateResp, error) {
	client := smsclient.NewCardTemplateServiceClient(m.cli.Conn())
	return client.AddCardTemplate(ctx, in, opts...)
}

func (m *defaultCardTemplateService) UpdateCardTemplate(ctx context.Context, in *UpdateCardTemplateReq, opts ...grpc.CallOption) (*UpdateCardTemplateResp, error) {
	client := smsclient.NewCardTemplateServiceClient(m.cli.Conn())
	return client.UpdateCardTemplate(ctx, in, opts...)
}

func (m *defaultCardTemplateService) QueryCardTemplateList(ctx context.Context, in *QueryCardTemplateListReq, opts ...grpc.CallOption) (*QueryCardTemplateListResp, error) {
	client := smsclient.NewCardTemplateServiceClient(m.cli.Conn())
	return client.QueryCardTemplateList(ctx, in, opts...)
}

func (m *defaultCardTemplateService) QueryCardTemplateDetail(ctx context.Context, in *QueryCardTemplateDetailReq, opts ...grpc.CallOption) (*QueryCardTemplateDetailResp, error) {
	client := smsclient.NewCardTemplateServiceClient(m.cli.Conn())
	return client.QueryCardTemplateDetail(ctx, in, opts...)
}

func (m *defaultCardTemplateService) UpdateCardTemplateStatus(ctx context.Context, in *UpdateCardTemplateStatusReq, opts ...grpc.CallOption) (*UpdateCardTemplateStatusResp, error) {
	client := smsclient.NewCardTemplateServiceClient(m.cli.Conn())
	return client.UpdateCardTemplateStatus(ctx, in, opts...)
}

func (m *defaultCardTemplateService) DeleteCardTemplate(ctx context.Context, in *DeleteCardTemplateReq, opts ...grpc.CallOption) (*DeleteCardTemplateResp, error) {
	client := smsclient.NewCardTemplateServiceClient(m.cli.Conn())
	return client.DeleteCardTemplate(ctx, in, opts...)
}

func (m *defaultCardTemplateService) CheckCardTemplateUsage(ctx context.Context, in *CheckCardTemplateUsageReq, opts ...grpc.CallOption) (*CheckCardTemplateUsageResp, error) {
	client := smsclient.NewCardTemplateServiceClient(m.cli.Conn())
	return client.CheckCardTemplateUsage(ctx, in, opts...)
}
