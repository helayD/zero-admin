package handler

import (
	"net/http"

	channelintegrationtemplatehandler "github.com/feihua/zero-admin/api/admin/internal/handler/sys/channel_integration_template"
	"github.com/feihua/zero-admin/api/admin/internal/svc"
	"github.com/zeromicro/go-zero/rest"
)

func RegisterExtraHandlers(server *rest.Server, serverCtx *svc.ServiceContext) {
	server.AddRoutes(
		rest.WithMiddlewares(
			[]rest.Middleware{serverCtx.CheckUrl},
			[]rest.Route{
				{
					Method:  http.MethodPost,
					Path:    "/createChannelIntegrationTemplate",
					Handler: channelintegrationtemplatehandler.CreateChannelIntegrationTemplateHandler(serverCtx),
				},
				{
					Method:  http.MethodPost,
					Path:    "/updateChannelIntegrationTemplate",
					Handler: channelintegrationtemplatehandler.UpdateChannelIntegrationTemplateHandler(serverCtx),
				},
				{
					Method:  http.MethodPost,
					Path:    "/updateChannelIntegrationTemplateStatus",
					Handler: channelintegrationtemplatehandler.UpdateChannelIntegrationTemplateStatusHandler(serverCtx),
				},
				{
					Method:  http.MethodGet,
					Path:    "/queryChannelIntegrationTemplateDetail",
					Handler: channelintegrationtemplatehandler.QueryChannelIntegrationTemplateDetailHandler(serverCtx),
				},
				{
					Method:  http.MethodGet,
					Path:    "/queryChannelIntegrationTemplateList",
					Handler: channelintegrationtemplatehandler.QueryChannelIntegrationTemplateListHandler(serverCtx),
				},
			}...,
		),
		rest.WithJwt(serverCtx.Config.Auth.AccessSecret),
		rest.WithPrefix("/api/sys/channelIntegrationTemplate"),
	)
}
