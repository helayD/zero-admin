package digital_card_asset

import (
	"net/http"

	"github.com/feihua/zero-admin/api/front/internal/logic/digital_card/digital_card_asset"
	"github.com/feihua/zero-admin/api/front/internal/svc"
	"github.com/feihua/zero-admin/api/front/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func QueryMyDigitalCardAssetDetailHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.QueryMyDigitalCardAssetDetailReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := digital_card_asset.NewQueryMyDigitalCardAssetDetailLogic(r.Context(), svcCtx)
		resp, err := l.QueryMyDigitalCardAssetDetail(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
