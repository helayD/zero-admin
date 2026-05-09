package handler

import (
	"net/http"

	cardtemplatehandler "github.com/feihua/zero-admin/api/admin/internal/handler/sms/card_template"
	digitalcardassethandler "github.com/feihua/zero-admin/api/admin/internal/handler/sms/digital_card_asset"
	digitalcardchainhandler "github.com/feihua/zero-admin/api/admin/internal/handler/sms/digital_card_chain"
	digitalcardphysicalfulfillmenthandler "github.com/feihua/zero-admin/api/admin/internal/handler/sms/digital_card_physical_fulfillment"
	productfulfillmentrulehandler "github.com/feihua/zero-admin/api/admin/internal/handler/sms/product_fulfillment_rule"
	channelintegrationtemplatehandler "github.com/feihua/zero-admin/api/admin/internal/handler/sys/channel_integration_template"
	systemconfighandler "github.com/feihua/zero-admin/api/admin/internal/handler/sys/system_config"
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
					Path:    "/querySystemConfig",
					Handler: systemconfighandler.QuerySystemConfigHandler(serverCtx),
				},
				{
					Method:  http.MethodPost,
					Path:    "/saveSystemConfig",
					Handler: systemconfighandler.SaveSystemConfigHandler(serverCtx),
				},
			}...,
		),
		rest.WithJwt(serverCtx.Config.Auth.AccessSecret),
		rest.WithPrefix("/api/sys/systemConfig"),
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

	server.AddRoutes(
		rest.WithMiddlewares(
			[]rest.Middleware{serverCtx.CheckUrl},
			[]rest.Route{
				{
					Method:  http.MethodPost,
					Path:    "/addProductFulfillmentRule",
					Handler: productfulfillmentrulehandler.AddProductFulfillmentRuleHandler(serverCtx),
				},
				{
					Method:  http.MethodPost,
					Path:    "/updateProductFulfillmentRule",
					Handler: productfulfillmentrulehandler.UpdateProductFulfillmentRuleHandler(serverCtx),
				},
				{
					Method:  http.MethodPost,
					Path:    "/updateProductFulfillmentRuleStatus",
					Handler: productfulfillmentrulehandler.UpdateProductFulfillmentRuleStatusHandler(serverCtx),
				},
				{
					Method:  http.MethodGet,
					Path:    "/deleteProductFulfillmentRule",
					Handler: productfulfillmentrulehandler.DeleteProductFulfillmentRuleHandler(serverCtx),
				},
				{
					Method:  http.MethodGet,
					Path:    "/queryProductFulfillmentRuleList",
					Handler: productfulfillmentrulehandler.QueryProductFulfillmentRuleListHandler(serverCtx),
				},
				{
					Method:  http.MethodGet,
					Path:    "/queryProductFulfillmentRuleDetail",
					Handler: productfulfillmentrulehandler.QueryProductFulfillmentRuleDetailHandler(serverCtx),
				},
				{
					Method:  http.MethodGet,
					Path:    "/checkProductFulfillmentRuleBinding",
					Handler: productfulfillmentrulehandler.CheckProductFulfillmentRuleBindingHandler(serverCtx),
				},
			}...,
		),
		rest.WithJwt(serverCtx.Config.Auth.AccessSecret),
		rest.WithPrefix("/api/sms/productFulfillmentRule"),
	)

	// 卡片模板（Story 10.10）
	server.AddRoutes(
		rest.WithMiddlewares(
			[]rest.Middleware{serverCtx.CheckUrl},
			[]rest.Route{
				{
					Method:  http.MethodPost,
					Path:    "/addCardTemplate",
					Handler: cardtemplatehandler.AddCardTemplateHandler(serverCtx),
				},
				{
					Method:  http.MethodPost,
					Path:    "/updateCardTemplate",
					Handler: cardtemplatehandler.UpdateCardTemplateHandler(serverCtx),
				},
				{
					Method:  http.MethodPost,
					Path:    "/updateCardTemplateStatus",
					Handler: cardtemplatehandler.UpdateCardTemplateStatusHandler(serverCtx),
				},
				{
					Method:  http.MethodGet,
					Path:    "/deleteCardTemplate",
					Handler: cardtemplatehandler.DeleteCardTemplateHandler(serverCtx),
				},
				{
					Method:  http.MethodGet,
					Path:    "/queryCardTemplateList",
					Handler: cardtemplatehandler.QueryCardTemplateListHandler(serverCtx),
				},
				{
					Method:  http.MethodGet,
					Path:    "/queryCardTemplateDetail",
					Handler: cardtemplatehandler.QueryCardTemplateDetailHandler(serverCtx),
				},
				{
					Method:  http.MethodGet,
					Path:    "/checkCardTemplateUsage",
					Handler: cardtemplatehandler.CheckCardTemplateUsageHandler(serverCtx),
				},
			}...,
		),
		rest.WithJwt(serverCtx.Config.Auth.AccessSecret),
		rest.WithPrefix("/api/sms/cardTemplate"),
	)
}
