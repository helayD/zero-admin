package handler

import (
	"net/http"

	digitalcardassethandler "github.com/feihua/zero-admin/api/admin/internal/handler/sms/digital_card_asset"
	digitalcardchainhandler "github.com/feihua/zero-admin/api/admin/internal/handler/sms/digital_card_chain"
	digitalcardphysicalfulfillmenthandler "github.com/feihua/zero-admin/api/admin/internal/handler/sms/digital_card_physical_fulfillment"
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

	server.AddRoutes(
		rest.WithMiddlewares(
			[]rest.Middleware{serverCtx.CheckUrl},
			[]rest.Route{
				{
					Method:  http.MethodGet,
					Path:    "/queryDigitalCardAssetList",
					Handler: digitalcardassethandler.QueryDigitalCardAssetListHandler(serverCtx),
				},
				{
					Method:  http.MethodGet,
					Path:    "/queryDigitalCardAssetDetail",
					Handler: digitalcardassethandler.QueryDigitalCardAssetDetailHandler(serverCtx),
				},
				{
					Method:  http.MethodGet,
					Path:    "/queryDigitalCardAssetLogs",
					Handler: digitalcardassethandler.QueryDigitalCardAssetLogsHandler(serverCtx),
				},
				{
					Method:  http.MethodPost,
					Path:    "/reviewDigitalCardAssetCompliance",
					Handler: digitalcardassethandler.ReviewDigitalCardAssetComplianceHandler(serverCtx),
				},
				{
					Method:  http.MethodPost,
					Path:    "/offlineDigitalCardAssetDisplay",
					Handler: digitalcardassethandler.OfflineDigitalCardAssetDisplayHandler(serverCtx),
				},
				{
					Method:  http.MethodPost,
					Path:    "/recycleDigitalCardAsset",
					Handler: digitalcardassethandler.RecycleDigitalCardAssetHandler(serverCtx),
				},
			}...,
		),
		rest.WithJwt(serverCtx.Config.Auth.AccessSecret),
		rest.WithPrefix("/api/sms/digitalCardAsset"),
	)

	server.AddRoutes(
		rest.WithMiddlewares(
			[]rest.Middleware{serverCtx.CheckUrl},
			[]rest.Route{
				{
					Method:  http.MethodGet,
					Path:    "/queryDigitalCardChainList",
					Handler: digitalcardchainhandler.QueryDigitalCardChainListHandler(serverCtx),
				},
				{
					Method:  http.MethodGet,
					Path:    "/queryDigitalCardChainDetail",
					Handler: digitalcardchainhandler.QueryDigitalCardChainDetailHandler(serverCtx),
				},
				{
					Method:  http.MethodGet,
					Path:    "/queryDigitalCardChainActions",
					Handler: digitalcardchainhandler.QueryDigitalCardChainActionsHandler(serverCtx),
				},
				{
					Method:  http.MethodPost,
					Path:    "/retryDigitalCardChain",
					Handler: digitalcardchainhandler.RetryDigitalCardChainHandler(serverCtx),
				},
				{
					Method:  http.MethodPost,
					Path:    "/freezeDigitalCardChain",
					Handler: digitalcardchainhandler.FreezeDigitalCardChainHandler(serverCtx),
				},
				{
					Method:  http.MethodPost,
					Path:    "/escalateDigitalCardChain",
					Handler: digitalcardchainhandler.EscalateDigitalCardChainHandler(serverCtx),
				},
			}...,
		),
		rest.WithJwt(serverCtx.Config.Auth.AccessSecret),
		rest.WithPrefix("/api/sms/digitalCardChain"),
	)

	server.AddRoutes(
		rest.WithMiddlewares(
			[]rest.Middleware{serverCtx.CheckUrl},
			[]rest.Route{
				{
					Method:  http.MethodGet,
					Path:    "/queryDigitalCardPhysicalFulfillmentList",
					Handler: digitalcardphysicalfulfillmenthandler.QueryDigitalCardPhysicalFulfillmentListHandler(serverCtx),
				},
				{
					Method:  http.MethodGet,
					Path:    "/queryDigitalCardPhysicalFulfillmentDetail",
					Handler: digitalcardphysicalfulfillmenthandler.QueryDigitalCardPhysicalFulfillmentDetailHandler(serverCtx),
				},
				{
					Method:  http.MethodPost,
					Path:    "/ensureDigitalCardPhysicalFulfillment",
					Handler: digitalcardphysicalfulfillmenthandler.EnsureDigitalCardPhysicalFulfillmentHandler(serverCtx),
				},
				{
					Method:  http.MethodPost,
					Path:    "/updateDigitalCardPhysicalProductionStatus",
					Handler: digitalcardphysicalfulfillmenthandler.UpdateDigitalCardPhysicalProductionStatusHandler(serverCtx),
				},
				{
					Method:  http.MethodPost,
					Path:    "/shipDigitalCardPhysicalFulfillment",
					Handler: digitalcardphysicalfulfillmenthandler.ShipDigitalCardPhysicalFulfillmentHandler(serverCtx),
				},
				{
					Method:  http.MethodPost,
					Path:    "/markDigitalCardPhysicalFulfillmentException",
					Handler: digitalcardphysicalfulfillmenthandler.MarkDigitalCardPhysicalFulfillmentExceptionHandler(serverCtx),
				},
				{
					Method:  http.MethodPost,
					Path:    "/requestDigitalCardPhysicalReissue",
					Handler: digitalcardphysicalfulfillmenthandler.RequestDigitalCardPhysicalReissueHandler(serverCtx),
				},
			}...,
		),
		rest.WithJwt(serverCtx.Config.Auth.AccessSecret),
		rest.WithPrefix("/api/sms/digitalCardPhysicalFulfillment"),
	)
}
