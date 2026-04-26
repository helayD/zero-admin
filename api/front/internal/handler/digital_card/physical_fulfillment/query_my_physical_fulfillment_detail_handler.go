package physical_fulfillment

import (
	"net/http"

	"github.com/feihua/zero-admin/api/front/internal/logic/digital_card/physical_fulfillment"
	"github.com/feihua/zero-admin/api/front/internal/svc"
	"github.com/feihua/zero-admin/api/front/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func QueryMyPhysicalFulfillmentDetailHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.QueryMyPhysicalFulfillmentDetailReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := physical_fulfillment.NewQueryMyPhysicalFulfillmentDetailLogic(r.Context(), svcCtx)
		resp, err := l.QueryMyPhysicalFulfillmentDetail(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
