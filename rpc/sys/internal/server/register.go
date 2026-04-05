package server

import (
	channelintegrationtemplateserviceServer "github.com/feihua/zero-admin/rpc/sys/internal/server/channelintegrationtemplateservice"
	"github.com/feihua/zero-admin/rpc/sys/internal/svc"
	"github.com/feihua/zero-admin/rpc/sys/sysclient"
	"google.golang.org/grpc"
)

func RegisterExtraServices(grpcServer *grpc.Server, svcCtx *svc.ServiceContext) {
	sysclient.RegisterChannelIntegrationTemplateServiceServer(grpcServer, channelintegrationtemplateserviceServer.NewChannelIntegrationTemplateServiceServer(svcCtx))
}
