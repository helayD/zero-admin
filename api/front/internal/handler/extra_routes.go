package handler

import (
	"net/http"

	digitalcardassethandler "github.com/feihua/zero-admin/api/front/internal/handler/digital_card/digital_card_asset"
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
}
