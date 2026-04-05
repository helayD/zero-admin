package server

import (
	"context"

	channelintegrationtemplateservicelogic "github.com/feihua/zero-admin/rpc/sys/internal/logic/channelintegrationtemplateservice"
	"github.com/feihua/zero-admin/rpc/sys/internal/svc"
	"github.com/feihua/zero-admin/rpc/sys/sysclient"
)

type ChannelIntegrationTemplateServiceServer struct {
	svcCtx *svc.ServiceContext
	sysclient.UnimplementedChannelIntegrationTemplateServiceServer
}

func NewChannelIntegrationTemplateServiceServer(svcCtx *svc.ServiceContext) *ChannelIntegrationTemplateServiceServer {
	return &ChannelIntegrationTemplateServiceServer{svcCtx: svcCtx}
}

func (s *ChannelIntegrationTemplateServiceServer) CreateChannelIntegrationTemplate(ctx context.Context, in *sysclient.CreateChannelIntegrationTemplateReq) (*sysclient.CreateChannelIntegrationTemplateResp, error) {
	l := channelintegrationtemplateservicelogic.NewCreateChannelIntegrationTemplateLogic(ctx, s.svcCtx)
	return l.CreateChannelIntegrationTemplate(in)
}

func (s *ChannelIntegrationTemplateServiceServer) UpdateChannelIntegrationTemplate(ctx context.Context, in *sysclient.UpdateChannelIntegrationTemplateReq) (*sysclient.UpdateChannelIntegrationTemplateResp, error) {
	l := channelintegrationtemplateservicelogic.NewUpdateChannelIntegrationTemplateLogic(ctx, s.svcCtx)
	return l.UpdateChannelIntegrationTemplate(in)
}

func (s *ChannelIntegrationTemplateServiceServer) UpdateChannelIntegrationTemplateStatus(ctx context.Context, in *sysclient.UpdateChannelIntegrationTemplateStatusReq) (*sysclient.UpdateChannelIntegrationTemplateStatusResp, error) {
	l := channelintegrationtemplateservicelogic.NewUpdateChannelIntegrationTemplateStatusLogic(ctx, s.svcCtx)
	return l.UpdateChannelIntegrationTemplateStatus(in)
}

func (s *ChannelIntegrationTemplateServiceServer) QueryChannelIntegrationTemplateDetail(ctx context.Context, in *sysclient.QueryChannelIntegrationTemplateDetailReq) (*sysclient.QueryChannelIntegrationTemplateDetailResp, error) {
	l := channelintegrationtemplateservicelogic.NewQueryChannelIntegrationTemplateDetailLogic(ctx, s.svcCtx)
	return l.QueryChannelIntegrationTemplateDetail(in)
}

func (s *ChannelIntegrationTemplateServiceServer) QueryChannelIntegrationTemplateList(ctx context.Context, in *sysclient.QueryChannelIntegrationTemplateListReq) (*sysclient.QueryChannelIntegrationTemplateListResp, error) {
	l := channelintegrationtemplateservicelogic.NewQueryChannelIntegrationTemplateListLogic(ctx, s.svcCtx)
	return l.QueryChannelIntegrationTemplateList(in)
}
