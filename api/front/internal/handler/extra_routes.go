package handler

import (
	"net/http"

	digitalcardassethandler "github.com/feihua/zero-admin/api/front/internal/handler/digital_card/digital_card_asset"
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
		},
		rest.WithJwt(serverCtx.Config.Auth.AccessSecret),
		rest.WithPrefix("/api/digitalCard/asset"),
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
