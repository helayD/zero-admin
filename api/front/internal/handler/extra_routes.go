package handler

import (
	"net/http"

	digitalcardassethandler "github.com/feihua/zero-admin/api/front/internal/handler/digital_card/digital_card_asset"
	"github.com/feihua/zero-admin/api/front/internal/middleware"
	physicalfulfillmenthandler "github.com/feihua/zero-admin/api/front/internal/handler/digital_card/physical_fulfillment"
	membermessagehandler "github.com/feihua/zero-admin/api/front/internal/handler/member/message"
	"github.com/feihua/zero-admin/api/front/internal/svc"
	"github.com/zeromicro/go-zero/rest"
)

func RegisterExtraHandlers(server *rest.Server, serverCtx *svc.ServiceContext) {
	server.AddRoutes(
		[]rest.Route{
			{
				Method:  http.MethodGet,
				Path:    "/queryMyDigitalCardAssetList",
				Handler: digitalcardassethandler.QueryMyDigitalCardAssetListHandler(serverCtx),
			},
			{
				Method:  http.MethodGet,
				Path:    "/queryMyDigitalCardAssetDetail",
				Handler: digitalcardassethandler.QueryMyDigitalCardAssetDetailHandler(serverCtx),
			},
			{
				Method:  http.MethodPost,
				Path:    "/resolveTransferRecipient",
				Handler: digitalcardassethandler.ResolveDigitalCardTransferRecipientHandler(serverCtx),
			},
			{
				Method:  http.MethodPost,
				Path:    "/transferDigitalCardAsset",
				Handler: digitalcardassethandler.TransferDigitalCardAssetHandler(serverCtx),
			},
			{
				Method:  http.MethodPost,
				Path:    "/requestDigitalCardWithdraw",
				Handler: digitalcardassethandler.RequestDigitalCardWithdrawHandler(serverCtx),
			},
			{
				Method:  http.MethodPost,
				Path:    "/createRedemptionOrder",
				Handler: digitalcardassethandler.CreateRedemptionOrderHandler(serverCtx),
			},
			{
				Method:  http.MethodGet,
				Path:    "/queryRedemptionOrder",
				Handler: digitalcardassethandler.QueryRedemptionOrderHandler(serverCtx),
			},
			{
				Method:  http.MethodPost,
				Path:    "/generateShareLink",
				Handler: digitalcardassethandler.GenerateShareLinkHandler(serverCtx),
			},
		},
		rest.WithJwt(serverCtx.Config.Auth.AccessSecret),
		rest.WithPrefix("/api/digitalCard/asset"),
	)

	server.AddRoutes(
		[]rest.Route{
			{
				Method:  http.MethodPost,
				Path:    "/validateClaimToken",
				Handler: digitalcardassethandler.ValidateClaimTokenHandler(serverCtx),
			},
		},
		rest.WithPrefix("/api/digitalCard"),
	)

	server.AddRoutes(
		[]rest.Route{
			{
				Method:  http.MethodPost,
				Path:    "/claim",
				Handler: middleware.DigitalCardRateLimitMiddleware(serverCtx.Redis)(
					middleware.AbnormalDetectionMiddleware(serverCtx.Redis)(
						digitalcardassethandler.ClaimDigitalCardHandler(serverCtx),
					),
				),
			},
		},
		rest.WithJwt(serverCtx.Config.Auth.AccessSecret),
		rest.WithPrefix("/api/digitalCard"),
	)

	server.AddRoutes(
		[]rest.Route{
			{
				Method:  http.MethodGet,
				Path:    "/digitalCard/register",
				Handler: digitalcardassethandler.DigitalCardRegisterHintHandler(serverCtx),
			},
		},
		rest.WithPrefix("/h5"),
	)

	server.AddRoutes(
		[]rest.Route{
			{
				Method:  http.MethodGet,
				Path:    "/queryMyPhysicalFulfillmentDetail",
				Handler: physicalfulfillmenthandler.QueryMyPhysicalFulfillmentDetailHandler(serverCtx),
			},
			{
				Method:  http.MethodPost,
				Path:    "/confirmPhysicalFulfillmentAddress",
				Handler: physicalfulfillmenthandler.ConfirmPhysicalFulfillmentAddressHandler(serverCtx),
			},
			{
				Method:  http.MethodPost,
				Path:    "/confirmPhysicalCardReceipt",
				Handler: physicalfulfillmenthandler.ConfirmPhysicalCardReceiptHandler(serverCtx),
			},
			{
				Method:  http.MethodPost,
				Path:    "/confirmPhysicalFulfillmentShippingFee",
				Handler: physicalfulfillmenthandler.ConfirmPhysicalFulfillmentShippingFeeHandler(serverCtx),
			},
		},
		rest.WithJwt(serverCtx.Config.Auth.AccessSecret),
		rest.WithPrefix("/api/digitalCard/physicalFulfillment"),
	)

	server.AddRoutes(
		[]rest.Route{
			{
				Method:  http.MethodGet,
				Path:    "/list",
				Handler: membermessagehandler.MessageListHandler(serverCtx),
			},
			{
				Method:  http.MethodGet,
				Path:    "/unreadCount",
				Handler: membermessagehandler.QueryUnreadCountHandler(serverCtx),
			},
			{
				Method:  http.MethodPost,
				Path:    "/readAll",
				Handler: membermessagehandler.MarkAllMessagesReadHandler(serverCtx),
			},
			{
				Method:  http.MethodPost,
				Path:    "/read",
				Handler: membermessagehandler.MarkMessageReadHandler(serverCtx),
			},
			{
				Method:  http.MethodPost,
				Path:    "/delete",
				Handler: membermessagehandler.DeleteMessageHandler(serverCtx),
			},
			{
				Method:  http.MethodDelete,
				Path:    "/delete",
				Handler: membermessagehandler.DeleteMessageHandler(serverCtx),
			},
			{
				Method:  http.MethodGet,
				Path:    "/:id",
				Handler: membermessagehandler.MessageDetailHandler(serverCtx),
			},
			{
				Method:  http.MethodPost,
				Path:    "/:id/read",
				Handler: membermessagehandler.MarkMessageReadByPathHandler(serverCtx),
			},
			{
				Method:  http.MethodDelete,
				Path:    "/:id",
				Handler: membermessagehandler.DeleteMessageByPathHandler(serverCtx),
			},
		},
		rest.WithJwt(serverCtx.Config.Auth.AccessSecret),
		rest.WithPrefix("/api/member/message"),
	)
}
