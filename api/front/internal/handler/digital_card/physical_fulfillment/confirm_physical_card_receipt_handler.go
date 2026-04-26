package physical_fulfillment

import (
	"net/http"

	"github.com/feihua/zero-admin/api/front/internal/logic/digital_card/physical_fulfillment"
	"github.com/feihua/zero-admin/api/front/internal/svc"
	"github.com/feihua/zero-admin/api/front/internal/types"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func ConfirmPhysicalCardReceiptHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.ConfirmPhysicalCardReceiptReq
		if err := httpx.Parse(r, &req); err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
			return
		}

		l := physical_fulfillment.NewConfirmPhysicalCardReceiptLogic(r.Context(), svcCtx)
		resp, err := l.ConfirmPhysicalCardReceipt(&req)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
