package channelintegrationtemplateservice

import (
	"context"

	"github.com/feihua/zero-admin/rpc/sys/sysclient"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

type (
	ChannelIntegrationTemplateData             = sysclient.ChannelIntegrationTemplateData
	CreateChannelIntegrationTemplateReq        = sysclient.CreateChannelIntegrationTemplateReq
	CreateChannelIntegrationTemplateResp       = sysclient.CreateChannelIntegrationTemplateResp
	UpdateChannelIntegrationTemplateReq        = sysclient.UpdateChannelIntegrationTemplateReq
	UpdateChannelIntegrationTemplateResp       = sysclient.UpdateChannelIntegrationTemplateResp
	UpdateChannelIntegrationTemplateStatusReq  = sysclient.UpdateChannelIntegrationTemplateStatusReq
	UpdateChannelIntegrationTemplateStatusResp = sysclient.UpdateChannelIntegrationTemplateStatusResp
	QueryChannelIntegrationTemplateDetailReq   = sysclient.QueryChannelIntegrationTemplateDetailReq
	QueryChannelIntegrationTemplateDetailResp  = sysclient.QueryChannelIntegrationTemplateDetailResp
	QueryChannelIntegrationTemplateListReq     = sysclient.QueryChannelIntegrationTemplateListReq
	QueryChannelIntegrationTemplateListResp    = sysclient.QueryChannelIntegrationTemplateListResp

	ChannelIntegrationTemplateService interface {
		CreateChannelIntegrationTemplate(ctx context.Context, in *CreateChannelIntegrationTemplateReq, opts ...grpc.CallOption) (*CreateChannelIntegrationTemplateResp, error)
		UpdateChannelIntegrationTemplate(ctx context.Context, in *UpdateChannelIntegrationTemplateReq, opts ...grpc.CallOption) (*UpdateChannelIntegrationTemplateResp, error)
		UpdateChannelIntegrationTemplateStatus(ctx context.Context, in *UpdateChannelIntegrationTemplateStatusReq, opts ...grpc.CallOption) (*UpdateChannelIntegrationTemplateStatusResp, error)
		QueryChannelIntegrationTemplateDetail(ctx context.Context, in *QueryChannelIntegrationTemplateDetailReq, opts ...grpc.CallOption) (*QueryChannelIntegrationTemplateDetailResp, error)
		QueryChannelIntegrationTemplateList(ctx context.Context, in *QueryChannelIntegrationTemplateListReq, opts ...grpc.CallOption) (*QueryChannelIntegrationTemplateListResp, error)
	}

	defaultChannelIntegrationTemplateService struct {
		cli zrpc.Client
	}
)

func NewChannelIntegrationTemplateService(cli zrpc.Client) ChannelIntegrationTemplateService {
	return &defaultChannelIntegrationTemplateService{cli: cli}
}

func (m *defaultChannelIntegrationTemplateService) CreateChannelIntegrationTemplate(ctx context.Context, in *CreateChannelIntegrationTemplateReq, opts ...grpc.CallOption) (*CreateChannelIntegrationTemplateResp, error) {
	client := sysclient.NewChannelIntegrationTemplateServiceClient(m.cli.Conn())
	return client.CreateChannelIntegrationTemplate(ctx, in, opts...)
}

func (m *defaultChannelIntegrationTemplateService) UpdateChannelIntegrationTemplate(ctx context.Context, in *UpdateChannelIntegrationTemplateReq, opts ...grpc.CallOption) (*UpdateChannelIntegrationTemplateResp, error) {
	client := sysclient.NewChannelIntegrationTemplateServiceClient(m.cli.Conn())
	return client.UpdateChannelIntegrationTemplate(ctx, in, opts...)
}

func (m *defaultChannelIntegrationTemplateService) UpdateChannelIntegrationTemplateStatus(ctx context.Context, in *UpdateChannelIntegrationTemplateStatusReq, opts ...grpc.CallOption) (*UpdateChannelIntegrationTemplateStatusResp, error) {
	client := sysclient.NewChannelIntegrationTemplateServiceClient(m.cli.Conn())
	return client.UpdateChannelIntegrationTemplateStatus(ctx, in, opts...)
}

func (m *defaultChannelIntegrationTemplateService) QueryChannelIntegrationTemplateDetail(ctx context.Context, in *QueryChannelIntegrationTemplateDetailReq, opts ...grpc.CallOption) (*QueryChannelIntegrationTemplateDetailResp, error) {
	client := sysclient.NewChannelIntegrationTemplateServiceClient(m.cli.Conn())
	return client.QueryChannelIntegrationTemplateDetail(ctx, in, opts...)
}

func (m *defaultChannelIntegrationTemplateService) QueryChannelIntegrationTemplateList(ctx context.Context, in *QueryChannelIntegrationTemplateListReq, opts ...grpc.CallOption) (*QueryChannelIntegrationTemplateListResp, error) {
	client := sysclient.NewChannelIntegrationTemplateServiceClient(m.cli.Conn())
	return client.QueryChannelIntegrationTemplateList(ctx, in, opts...)
}
